//go:build integration

// DB-backed integration tests for the folder filing/reparent use case. Like the
// ACL integration tests they are excluded from the default `go test ./...` run
// by the `integration` build tag and require a live Postgres reachable at
// DATABASE_URL (migrated to at least 013). Every test runs inside a transaction
// that is always rolled back, so the target database is left untouched.
//
// These live in the external repository_test package (not package repository)
// because they drive usecase.FolderService, and usecase imports repository — an
// internal test would form an import cycle. The repository is wired to the test
// transaction via postgres.NewFolderRepositoryFromDBTX (integration-only).
package postgres_test

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"

	postgres "github.com/docvault/backend/internal/document/adapter/postgres"
	"github.com/docvault/backend/internal/migrate"
	"github.com/docvault/backend/internal/usecase"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const folderDefaultDatabaseURL = "postgresql://docvault:docvault_dev@localhost:5432/docvault"

func folderDatabaseURL() string {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	return folderDefaultDatabaseURL
}

// folderTenantFixture holds the tenant/org/user seeded for a folder test.
type folderTenantFixture struct {
	tenantID string
	orgID    string
	userID   string
}

// folderQueryID runs a RETURNING/SELECT expected to yield a single uuid.
func folderQueryID(t *testing.T, tx pgx.Tx, sql string, args ...interface{}) string {
	t.Helper()
	var id string
	if err := tx.QueryRow(context.Background(), sql, args...).Scan(&id); err != nil {
		t.Fatalf("query id failed: %v\nSQL: %s", err, sql)
	}
	return id
}

// folderSeedBase creates a tenant, org, and a user, returning their ids.
func folderSeedBase(t *testing.T, tx pgx.Tx) folderTenantFixture {
	t.Helper()
	var f folderTenantFixture
	f.tenantID = folderQueryID(t, tx,
		`INSERT INTO tenants (name, plan) VALUES ('folder-test', 'business') RETURNING id`)
	f.orgID = folderQueryID(t, tx,
		`INSERT INTO organizations (tenant_id, name) VALUES ($1, 'folder-org') RETURNING id`, f.tenantID)
	f.userID = folderQueryID(t, tx,
		`INSERT INTO users (tenant_id, email, display_name) VALUES ($1, 'u@folder.test', 'User') RETURNING id`, f.tenantID)
	return f
}

// withFolderTx opens a transaction, runs fn, and always rolls back so the DB
// stays clean regardless of assertions.
func withFolderTx(t *testing.T, fn func(tx pgx.Tx)) {
	t.Helper()
	ctx := context.Background()

	if err := migrate.Run(ctx, folderDatabaseURL()); err != nil {
		t.Fatalf("migrate.Run: %v", err)
	}

	pool, err := pgxpool.New(ctx, folderDatabaseURL())
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	defer pool.Close()

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("pool.Begin: %v", err)
	}
	defer func() {
		if rbErr := tx.Rollback(ctx); rbErr != nil && rbErr != pgx.ErrTxClosed {
			t.Errorf("rollback: %v", rbErr)
		}
	}()

	fn(tx)
}

// folderServiceForTx wires the real folder repository to the test transaction so
// the production code path (sqldb queries) is exercised end to end. The ACL repo
// is nil; none of the folder paths under test touch it.
func folderServiceForTx(tx pgx.Tx) *usecase.FolderService {
	repo := postgres.NewFolderRepositoryFromDBTX(tx)
	return usecase.NewFolderService(repo, nil)
}

// folderPoolQuerier is the subset of the pgx query surface shared by *pgxpool.Pool
// and pgx.Tx, letting the helpers below run against committed rows on a pool.
type folderPoolQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
}

// withFolderPool opens a migrated pool and seeds a tenant/org/user with REAL
// committed rows under a freshly generated tenant id, runs fn against a
// pool-backed FolderService (whose Reparent opens its own advisory-locked
// transaction and commits), and unconditionally deletes the seeded rows
// afterwards. Reparent cannot be exercised inside a single rolled-back
// transaction because the advisory lock + real commits require a separate
// transaction per call; this committed-rows-with-cleanup model is the
// alternative. The tenant id is unique per call so concurrent test runs and the
// advisory lock keyed on it never collide.
func withFolderPool(t *testing.T, fn func(pool *pgxpool.Pool, svc *usecase.FolderService, f folderTenantFixture)) {
	t.Helper()
	ctx := context.Background()

	if err := migrate.Run(ctx, folderDatabaseURL()); err != nil {
		t.Fatalf("migrate.Run: %v", err)
	}

	pool, err := pgxpool.New(ctx, folderDatabaseURL())
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	defer pool.Close()

	f := folderSeedBaseCommitted(t, pool)
	defer folderCleanupCommitted(t, pool, f)

	repo := postgres.NewFolderRepositoryFromPool(pool)
	svc := usecase.NewFolderService(repo, nil)
	fn(pool, svc, f)
}

// folderSeedBaseCommitted creates a tenant, org, and user with committed rows on
// the pool, returning their ids. Mirrors folderSeedBase but commits.
func folderSeedBaseCommitted(t *testing.T, q folderPoolQuerier) folderTenantFixture {
	t.Helper()
	ctx := context.Background()
	var f folderTenantFixture
	scan := func(sql string, args ...interface{}) string {
		var id string
		if err := q.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
			t.Fatalf("seed query failed: %v\nSQL: %s", err, sql)
		}
		return id
	}
	f.tenantID = scan(`INSERT INTO tenants (name, plan) VALUES ('folder-reparent-test', 'business') RETURNING id`)
	f.orgID = scan(`INSERT INTO organizations (tenant_id, name) VALUES ($1, 'folder-org') RETURNING id`, f.tenantID)
	f.userID = scan(`INSERT INTO users (tenant_id, email, display_name) VALUES ($1, $2, 'User') RETURNING id`,
		f.tenantID, "u+"+f.tenantID+"@folder.test")
	return f
}

// folderCleanupCommitted removes every row seeded under the throwaway tenant so
// the live database is left untouched. folders cascade-clean here explicitly.
func folderCleanupCommitted(t *testing.T, q folderPoolQuerier, f folderTenantFixture) {
	t.Helper()
	ctx := context.Background()
	for _, sql := range []string{
		`DELETE FROM folders WHERE tenant_id = $1`,
		`DELETE FROM users WHERE tenant_id = $1`,
		`DELETE FROM organizations WHERE tenant_id = $1`,
		`DELETE FROM tenants WHERE id = $1`,
	} {
		if _, err := q.Exec(ctx, sql, f.tenantID); err != nil {
			t.Errorf("cleanup %q: %v", sql, err)
		}
	}
}

// createFolderRow inserts a single folder with committed visibility and returns
// its id. parentID may be nil for a root folder.
func createFolderRow(t *testing.T, q folderPoolQuerier, f folderTenantFixture, name string, parentID *string) string {
	t.Helper()
	var id string
	if err := q.QueryRow(context.Background(),
		`INSERT INTO folders (id, tenant_id, org_id, parent_id, name, created_by, created_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, NOW()) RETURNING id`,
		f.tenantID, f.orgID, parentID, name, f.userID).Scan(&id); err != nil {
		t.Fatalf("insert folder %q: %v", name, err)
	}
	return id
}

// folderParentPool returns the parent_id of a folder via the pool (empty string
// when NULL).
func folderParentPool(t *testing.T, q folderPoolQuerier, folderID string) string {
	t.Helper()
	var parent *string
	if err := q.QueryRow(context.Background(),
		`SELECT parent_id FROM folders WHERE id = $1`, folderID).Scan(&parent); err != nil {
		t.Fatalf("read parent_id: %v", err)
	}
	if parent == nil {
		return ""
	}
	return *parent
}

// countFolders returns how many folders exist for the tenant/org.
func countFolders(t *testing.T, tx pgx.Tx, tenantID, orgID string) int {
	t.Helper()
	var n int
	if err := tx.QueryRow(context.Background(),
		`SELECT count(*) FROM folders WHERE tenant_id = $1 AND org_id = $2`,
		tenantID, orgID).Scan(&n); err != nil {
		t.Fatalf("count folders: %v", err)
	}
	return n
}

// folderParent returns the parent_id of a folder (empty string when NULL).
func folderParent(t *testing.T, tx pgx.Tx, folderID string) string {
	t.Helper()
	var parent *string
	if err := tx.QueryRow(context.Background(),
		`SELECT parent_id FROM folders WHERE id = $1`, folderID).Scan(&parent); err != nil {
		t.Fatalf("read parent_id: %v", err)
	}
	if parent == nil {
		return ""
	}
	return *parent
}

// TestEnsureFolderPath_CreatesNestedPathIdempotently asserts that EnsureFolderPath
// find-or-creates each segment of a nested path and is idempotent: a second call
// with the same path returns the same leaf id and creates no duplicate folders.
func TestEnsureFolderPath_CreatesNestedPathIdempotently(t *testing.T) {
	withFolderTx(t, func(tx pgx.Tx) {
		f := folderSeedBase(t, tx)
		svc := folderServiceForTx(tx)
		ctx := context.Background()

		// Nested path per the "/"-joined contract: split, trim, skip empties.
		segments := usecase.SplitFolderPath(" Clients / Acme / Invoices ")
		if len(segments) != 3 {
			t.Fatalf("SplitFolderPath normalized to %#v, want 3 segments", segments)
		}

		leaf1, err := svc.EnsureFolderPath(ctx, f.tenantID, f.orgID, f.userID, segments)
		if err != nil {
			t.Fatalf("EnsureFolderPath (first): %v", err)
		}
		if leaf1 == "" {
			t.Fatal("EnsureFolderPath returned an empty leaf id")
		}

		// Three folders were created: Clients, Acme, Invoices.
		if got := countFolders(t, tx, f.tenantID, f.orgID); got != 3 {
			t.Fatalf("after first EnsureFolderPath folder count = %d, want 3", got)
		}

		// Leaf's parent chain reflects the nesting (leaf is "Invoices").
		acmeID := folderParent(t, tx, leaf1)
		if acmeID == "" {
			t.Fatal("leaf folder has no parent; nesting was not created")
		}
		clientsID := folderParent(t, tx, acmeID)
		if clientsID == "" {
			t.Fatal("middle folder has no parent; nesting was not created")
		}
		if folderParent(t, tx, clientsID) != "" {
			t.Fatal("root segment should have a NULL parent")
		}

		// Idempotency: a second identical call returns the same leaf and adds nothing.
		leaf2, err := svc.EnsureFolderPath(ctx, f.tenantID, f.orgID, f.userID, segments)
		if err != nil {
			t.Fatalf("EnsureFolderPath (second): %v", err)
		}
		if leaf2 != leaf1 {
			t.Fatalf("EnsureFolderPath not idempotent: leaf2 = %s, leaf1 = %s", leaf2, leaf1)
		}
		if got := countFolders(t, tx, f.tenantID, f.orgID); got != 3 {
			t.Fatalf("after second EnsureFolderPath folder count = %d, want 3 (no duplicates)", got)
		}

		// A path sharing the "Clients" prefix reuses it and adds only the new leaf.
		other := usecase.SplitFolderPath("Clients/Globex")
		otherLeaf, err := svc.EnsureFolderPath(ctx, f.tenantID, f.orgID, f.userID, other)
		if err != nil {
			t.Fatalf("EnsureFolderPath (shared prefix): %v", err)
		}
		if folderParent(t, tx, otherLeaf) != clientsID {
			t.Fatal("shared-prefix path did not reuse the existing 'Clients' folder")
		}
		if got := countFolders(t, tx, f.tenantID, f.orgID); got != 4 {
			t.Fatalf("after shared-prefix EnsureFolderPath folder count = %d, want 4", got)
		}
	})
}

// TestReparent_GuardsCyclesAndAcceptsValidMove asserts the reparent guards: a
// self-parent is rejected, a descendant-as-parent (cycle) is rejected, and a
// valid move under an unrelated folder is accepted and persisted.
func TestReparent_GuardsCyclesAndAcceptsValidMove(t *testing.T) {
	// Uses committed rows on a pool because Reparent now opens its own
	// advisory-locked transaction and commits; a single rolled-back tx cannot
	// host it.
	withFolderPool(t, func(pool *pgxpool.Pool, svc *usecase.FolderService, f folderTenantFixture) {
		ctx := context.Background()

		// Tree: A (root) -> B (child of A) -> C (child of B); plus D at root.
		a := createFolderRow(t, pool, f, "A", nil)
		b := createFolderRow(t, pool, f, "B", &a)
		c := createFolderRow(t, pool, f, "C", &b)
		d := createFolderRow(t, pool, f, "D", nil)

		// Self-parent: moving A under A is rejected and does not mutate A.
		if err := svc.Reparent(ctx, f.tenantID, f.orgID, a, &a); !errors.Is(err, usecase.ErrFolderSelfParent) {
			t.Fatalf("Reparent self-parent err = %v, want ErrFolderSelfParent", err)
		}
		if folderParentPool(t, pool, a) != "" {
			t.Fatal("self-parent reparent must not change A's parent")
		}

		// Cycle: moving A under C (a descendant of A) is rejected.
		if err := svc.Reparent(ctx, f.tenantID, f.orgID, a, &c); !errors.Is(err, usecase.ErrFolderCycle) {
			t.Fatalf("Reparent descendant-as-parent err = %v, want ErrFolderCycle", err)
		}
		if folderParentPool(t, pool, a) != "" {
			t.Fatal("cycle-rejected reparent must not change A's parent")
		}
		// And the direct-child cycle case: moving A under B (also a descendant).
		if err := svc.Reparent(ctx, f.tenantID, f.orgID, a, &b); !errors.Is(err, usecase.ErrFolderCycle) {
			t.Fatalf("Reparent direct-descendant-as-parent err = %v, want ErrFolderCycle", err)
		}

		// Valid move: B (with its subtree) under the unrelated D is accepted.
		if err := svc.Reparent(ctx, f.tenantID, f.orgID, b, &d); err != nil {
			t.Fatalf("valid Reparent err = %v, want nil", err)
		}
		if got := folderParentPool(t, pool, b); got != d {
			t.Fatalf("after valid move B.parent = %s, want %s (D)", got, d)
		}
		// C still hangs under B (subtree moved with it); no folders were lost.
		if got := folderParentPool(t, pool, c); got != b {
			t.Fatalf("C.parent = %s, want %s (B) after moving B's subtree", got, b)
		}

		// Move-to-root: B back to root always succeeds (NULL parent).
		if err := svc.Reparent(ctx, f.tenantID, f.orgID, b, nil); err != nil {
			t.Fatalf("move-to-root Reparent err = %v, want nil", err)
		}
		if got := folderParentPool(t, pool, b); got != "" {
			t.Fatalf("after move-to-root B.parent = %q, want root (NULL)", got)
		}
	})
}

// TestReparent_RejectsUnknownTargetParent asserts the target parent must exist
// in the caller's tenant/org.
func TestReparent_RejectsUnknownTargetParent(t *testing.T) {
	withFolderPool(t, func(pool *pgxpool.Pool, svc *usecase.FolderService, f folderTenantFixture) {
		ctx := context.Background()

		a := createFolderRow(t, pool, f, "A", nil)

		missing := "00000000-0000-0000-0000-000000000000"
		if err := svc.Reparent(ctx, f.tenantID, f.orgID, a, &missing); !errors.Is(err, usecase.ErrTargetParentNotFound) {
			t.Fatalf("Reparent unknown target err = %v, want ErrTargetParentNotFound", err)
		}
	})
}

// TestReparent_ConcurrentOppositeMovesNeverCreateCycle is the regression test for
// the cycle race. With two sibling root folders A and B, it launches two
// goroutines that start together and call, respectively, Reparent(A under B) and
// Reparent(B under A). Before the per-tenant advisory lock both calls could pass
// their in-statement cycle check on a disjoint row and commit, leaving
// A.parent==B AND B.parent==A — a cycle. After serialization exactly one
// relationship is established and the other call is rejected with ErrFolderCycle.
//
// It runs several iterations to exercise the race. Because Reparent commits real
// rows, A and B are recreated fresh at root each iteration under a throwaway
// tenant that is deleted at the end.
func TestReparent_ConcurrentOppositeMovesNeverCreateCycle(t *testing.T) {
	withFolderPool(t, func(pool *pgxpool.Pool, svc *usecase.FolderService, f folderTenantFixture) {
		ctx := context.Background()

		const iterations = 8
		for i := 0; i < iterations; i++ {
			// Fresh siblings at root for this iteration.
			a := createFolderRow(t, pool, f, "A", nil)
			b := createFolderRow(t, pool, f, "B", nil)

			var wg sync.WaitGroup
			start := make(chan struct{})
			errs := make([]error, 2)

			wg.Add(2)
			// Goroutine 0: move A under B.
			go func() {
				defer wg.Done()
				<-start
				errs[0] = svc.Reparent(ctx, f.tenantID, f.orgID, a, &b)
			}()
			// Goroutine 1: move B under A.
			go func() {
				defer wg.Done()
				<-start
				errs[1] = svc.Reparent(ctx, f.tenantID, f.orgID, b, &a)
			}()

			close(start) // release both goroutines at once
			wg.Wait()

			aParent := folderParentPool(t, pool, a)
			bParent := folderParentPool(t, pool, b)

			// Core invariant: never a 2-cycle.
			if aParent == b && bParent == a {
				t.Fatalf("iteration %d: CYCLE created: A.parent==B and B.parent==A (A=%s B=%s)", i, a, b)
			}

			// Exactly one relationship established: one call succeeded, the other
			// was rejected as a cycle (or both rejected, which is also acyclic).
			rejected := 0
			for _, e := range errs {
				if e == nil {
					continue
				}
				if !errors.Is(e, usecase.ErrFolderCycle) {
					t.Fatalf("iteration %d: unexpected Reparent error = %v, want nil or ErrFolderCycle", i, e)
				}
				rejected++
			}
			if rejected == 0 {
				t.Fatalf("iteration %d: both reparents succeeded; expected at least one ErrFolderCycle rejection (A.parent=%q B.parent=%q)", i, aParent, bParent)
			}

			// At most one parent link should exist between A and B.
			links := 0
			if aParent == b {
				links++
			}
			if bParent == a {
				links++
			}
			if links > 1 {
				t.Fatalf("iteration %d: %d A<->B parent links, want <=1 (acyclic)", i, links)
			}

			// Reset for the next iteration: detach both to root, then delete.
			if _, err := pool.Exec(ctx,
				`UPDATE folders SET parent_id = NULL WHERE id = ANY($1)`,
				[]string{a, b}); err != nil {
				t.Fatalf("iteration %d: reset parents: %v", i, err)
			}
			if _, err := pool.Exec(ctx,
				`DELETE FROM folders WHERE id = ANY($1)`,
				[]string{a, b}); err != nil {
				t.Fatalf("iteration %d: delete A,B: %v", i, err)
			}
		}
	})
}
