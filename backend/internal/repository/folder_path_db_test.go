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
// transaction via repository.NewFolderRepositoryFromDBTX (integration-only).
package repository_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/docvault/backend/internal/migrate"
	"github.com/docvault/backend/internal/repository"
	"github.com/docvault/backend/internal/usecase"
	"github.com/jackc/pgx/v5"
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
	repo := repository.NewFolderRepositoryFromDBTX(tx)
	return usecase.NewFolderService(repo, nil)
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
	withFolderTx(t, func(tx pgx.Tx) {
		f := folderSeedBase(t, tx)
		svc := folderServiceForTx(tx)
		ctx := context.Background()

		// Tree: A (root) -> B (child of A) -> C (child of B); plus D at root.
		a, err := svc.EnsureFolderPath(ctx, f.tenantID, f.orgID, f.userID, []string{"A"})
		if err != nil {
			t.Fatalf("seed A: %v", err)
		}
		b, err := svc.EnsureFolderPath(ctx, f.tenantID, f.orgID, f.userID, []string{"A", "B"})
		if err != nil {
			t.Fatalf("seed B: %v", err)
		}
		c, err := svc.EnsureFolderPath(ctx, f.tenantID, f.orgID, f.userID, []string{"A", "B", "C"})
		if err != nil {
			t.Fatalf("seed C: %v", err)
		}
		d, err := svc.EnsureFolderPath(ctx, f.tenantID, f.orgID, f.userID, []string{"D"})
		if err != nil {
			t.Fatalf("seed D: %v", err)
		}

		// Self-parent: moving A under A is rejected and does not mutate A.
		if err := svc.Reparent(ctx, f.tenantID, f.orgID, a, &a); !errors.Is(err, usecase.ErrFolderSelfParent) {
			t.Fatalf("Reparent self-parent err = %v, want ErrFolderSelfParent", err)
		}
		if folderParent(t, tx, a) != "" {
			t.Fatal("self-parent reparent must not change A's parent")
		}

		// Cycle: moving A under C (a descendant of A) is rejected.
		if err := svc.Reparent(ctx, f.tenantID, f.orgID, a, &c); !errors.Is(err, usecase.ErrFolderCycle) {
			t.Fatalf("Reparent descendant-as-parent err = %v, want ErrFolderCycle", err)
		}
		if folderParent(t, tx, a) != "" {
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
		if got := folderParent(t, tx, b); got != d {
			t.Fatalf("after valid move B.parent = %s, want %s (D)", got, d)
		}
		// C still hangs under B (subtree moved with it); no folders were lost.
		if got := folderParent(t, tx, c); got != b {
			t.Fatalf("C.parent = %s, want %s (B) after moving B's subtree", got, b)
		}
		if got := countFolders(t, tx, f.tenantID, f.orgID); got != 4 {
			t.Fatalf("folder count = %d after reparent, want 4 (A,B,C,D preserved)", got)
		}
	})
}

// TestReparent_RejectsUnknownTargetParent asserts the target parent must exist
// in the caller's tenant/org.
func TestReparent_RejectsUnknownTargetParent(t *testing.T) {
	withFolderTx(t, func(tx pgx.Tx) {
		f := folderSeedBase(t, tx)
		svc := folderServiceForTx(tx)
		ctx := context.Background()

		a, err := svc.EnsureFolderPath(ctx, f.tenantID, f.orgID, f.userID, []string{"A"})
		if err != nil {
			t.Fatalf("seed A: %v", err)
		}

		missing := "00000000-0000-0000-0000-000000000000"
		if err := svc.Reparent(ctx, f.tenantID, f.orgID, a, &missing); !errors.Is(err, usecase.ErrTargetParentNotFound) {
			t.Fatalf("Reparent unknown target err = %v, want ErrTargetParentNotFound", err)
		}
	})
}
