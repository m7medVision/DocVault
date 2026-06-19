//go:build integration

// DB-backed integration tests for transactional self-registration. Unlike the
// rolled-back visibility tests, RegisterWorkspace commits, so each test uses
// unique identifiers and deletes the rows it created in t.Cleanup.
package repository

import (
	"context"
	"testing"
	"time"

	"github.com/docvault/backend/internal/migrate"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func newRegistrationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()
	if err := migrate.Run(ctx, databaseURL()); err != nil {
		t.Fatalf("migrate.Run: %v", err)
	}
	pool, err := pgxpool.New(ctx, databaseURL())
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// cleanupWorkspace removes every row a RegisterWorkspace call may have created,
// in dependency order, so a committed test leaves the database as it found it.
func cleanupWorkspace(t *testing.T, pool *pgxpool.Pool, tenantID string) {
	t.Helper()
	t.Cleanup(func() {
		ctx := context.Background()
		stmts := []string{
			`DELETE FROM memberships WHERE org_id IN (SELECT id FROM organizations WHERE tenant_id = $1)`,
			`DELETE FROM users WHERE tenant_id = $1`,
			`DELETE FROM organizations WHERE tenant_id = $1`,
			`DELETE FROM tenants WHERE id = $1`,
		}
		for _, s := range stmts {
			if _, err := pool.Exec(ctx, s, tenantID); err != nil {
				t.Errorf("cleanup %q: %v", s, err)
			}
		}
	})
}

func newWorkspaceParams() RegisterWorkspaceParams {
	suffix := uuid.NewString()
	return RegisterWorkspaceParams{
		TenantID:     uuid.NewString(),
		TenantName:   "reg-tenant",
		OrgID:        uuid.NewString(),
		OrgName:      "reg-org",
		UserID:       uuid.NewString(),
		Email:        "reg-" + suffix + "@reg.test",
		PasswordHash: "hash",
		DisplayName:  "Reg User",
		Locale:       "en",
		MembershipID: uuid.NewString(),
		CreatedAt:    time.Now().UTC(),
	}
}

func TestRegisterWorkspace_CreatesTenantOrgUserAndMembership(t *testing.T) {
	pool := newRegistrationPool(t)
	repo := NewRegistrationRepository(pool)
	ctx := context.Background()

	p := newWorkspaceParams()
	cleanupWorkspace(t, pool, p.TenantID)

	if err := repo.RegisterWorkspace(ctx, p); err != nil {
		t.Fatalf("RegisterWorkspace: %v", err)
	}

	count := func(sql string, args ...interface{}) int {
		var n int
		if err := pool.QueryRow(ctx, sql, args...).Scan(&n); err != nil {
			t.Fatalf("count: %v\nSQL: %s", err, sql)
		}
		return n
	}

	if n := count(`SELECT count(*) FROM tenants WHERE id = $1`, p.TenantID); n != 1 {
		t.Fatalf("tenants = %d, want 1", n)
	}
	if n := count(`SELECT count(*) FROM organizations WHERE id = $1 AND tenant_id = $2`, p.OrgID, p.TenantID); n != 1 {
		t.Fatalf("organizations = %d, want 1", n)
	}
	if n := count(`SELECT count(*) FROM users WHERE id = $1 AND tenant_id = $2 AND email = $3`, p.UserID, p.TenantID, p.Email); n != 1 {
		t.Fatalf("users = %d, want 1", n)
	}
	if n := count(`SELECT count(*) FROM memberships WHERE id = $1 AND user_id = $2 AND org_id = $3 AND role = 'owner'`, p.MembershipID, p.UserID, p.OrgID); n != 1 {
		t.Fatalf("owner memberships = %d, want 1", n)
	}
}

func TestRegisterWorkspace_FailedUserInsertRollsBackTheWholeWorkspace(t *testing.T) {
	pool := newRegistrationPool(t)
	repo := NewRegistrationRepository(pool)
	ctx := context.Background()

	first := newWorkspaceParams()
	cleanupWorkspace(t, pool, first.TenantID)
	if err := repo.RegisterWorkspace(ctx, first); err != nil {
		t.Fatalf("first RegisterWorkspace: %v", err)
	}

	// A second registration that reuses the first user's id fails on the user
	// insert (duplicate primary key), which must roll back the tenant and org
	// created earlier in the same transaction — no half-provisioned workspace.
	second := newWorkspaceParams()
	second.UserID = first.UserID
	cleanupWorkspace(t, pool, second.TenantID)

	if err := repo.RegisterWorkspace(ctx, second); err == nil {
		t.Fatal("expected duplicate-user registration to fail, got nil")
	}

	var n int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM tenants WHERE id = $1`, second.TenantID).Scan(&n); err != nil {
		t.Fatalf("count tenants: %v", err)
	}
	if n != 0 {
		t.Fatalf("rolled-back tenant still present: count = %d, want 0", n)
	}
}
