//go:build integration

// Package repository DB-backed integration tests for per-folder read
// visibility (the IsFolderVisible seam) and folder-index access control.
//
// These tests mirror the document-visibility suite in acl_db_test.go: they are
// excluded from the default `go test ./...` run by the `integration` build tag,
// require a live Postgres reachable at DATABASE_URL (migrated to at least 014),
// and run inside a transaction that is always rolled back, so the target
// database is left untouched. They reuse seedBase / withTx / aclRepoForTx /
// exec / queryID from acl_db_test.go.
package repository

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

// mustFolderVisible evaluates IsFolderVisible and fails the test on a wiring
// error so each case asserts only the boolean visibility decision.
func mustFolderVisible(t *testing.T, repo *aclRepository, params FolderVisibilityParams) bool {
	t.Helper()
	visible, err := repo.IsFolderVisible(context.Background(), params)
	if err != nil {
		t.Fatalf("IsFolderVisible error: %v", err)
	}
	return visible
}

// withFolderUser returns a copy of params with the user identity and admin flag
// set (the folder analogue of withUser in acl_db_test.go).
func withFolderUser(params FolderVisibilityParams, userID string, isAdmin bool) FolderVisibilityParams {
	params.UserID = userID
	params.IsAdmin = isAdmin
	return params
}

// TestFolderVisibility_RestrictedGrantAndAdminBypass covers the core folder
// rule: a RESTRICTED folder created (created_by) by userA is hidden from a
// non-granted member, visible to its owner, always visible to is_admin=true,
// and visible to a non-granted member once a read grant is added on the folder.
func TestFolderVisibility_RestrictedGrantAndAdminBypass(t *testing.T) {
	withTx(t, func(tx pgx.Tx) {
		f := seedBase(t, tx)

		// userA owns (created_by) a restricted folder.
		f.folderID = queryID(t, tx,
			`INSERT INTO folders (tenant_id, org_id, name, created_by, is_restricted)
			 VALUES ($1, $2, 'Restricted Folder', $3, true) RETURNING id`,
			f.tenantID, f.orgID, f.userA)

		repo := aclRepoForTx(tx)
		base := FolderVisibilityParams{TenantID: f.tenantID, OrgID: f.orgID, FolderID: f.folderID}

		// userB: not owner, not granted, not admin -> NOT visible.
		if mustFolderVisible(t, repo, withFolderUser(base, f.userB, false)) {
			t.Fatal("userB should not see a restricted folder they do not own and were not granted")
		}

		// owner userA -> visible (created_by match).
		if !mustFolderVisible(t, repo, withFolderUser(base, f.userA, false)) {
			t.Fatal("owner userA should see their own restricted folder")
		}

		// is_admin bypass for userB -> visible.
		if !mustFolderVisible(t, repo, withFolderUser(base, f.userB, true)) {
			t.Fatal("is_admin=true should bypass folder restriction")
		}

		// Grant read to userB on the folder -> visible.
		grantID, err := repo.CreateGrant(context.Background(), CreateGrantParams{
			TenantID: f.tenantID, OrgID: f.orgID,
			ResourceType: "folder", ResourceID: f.folderID,
			PrincipalType: "user", PrincipalID: f.userB, Permission: "read",
		})
		if err != nil {
			t.Fatalf("CreateGrant (folder read): %v", err)
		}
		if grantID == "" {
			t.Fatal("CreateGrant returned empty id for a new folder grant")
		}
		if !mustFolderVisible(t, repo, withFolderUser(base, f.userB, false)) {
			t.Fatal("userB should see the folder after a read grant")
		}
	})
}

// TestFolderVisibility_RestrictionCascade asserts a child folder under a
// RESTRICTED parent is hidden from a non-granted member (the restriction
// cascades down the recursive ancestor chain even though the child itself is
// not restricted), and a read grant on the restricted parent restores
// visibility of the child.
func TestFolderVisibility_RestrictionCascade(t *testing.T) {
	withTx(t, func(tx pgx.Tx) {
		f := seedBase(t, tx)

		// Two-level tree: parent (restricted) -> child (not restricted).
		parentFolder := queryID(t, tx,
			`INSERT INTO folders (tenant_id, org_id, name, created_by, is_restricted)
			 VALUES ($1, $2, 'Restricted Parent', $3, true) RETURNING id`,
			f.tenantID, f.orgID, f.userA)
		childFolder := queryID(t, tx,
			`INSERT INTO folders (tenant_id, org_id, parent_id, name, created_by, is_restricted)
			 VALUES ($1, $2, $3, 'Open Child', $4, false) RETURNING id`,
			f.tenantID, f.orgID, parentFolder, f.userA)

		repo := aclRepoForTx(tx)
		base := FolderVisibilityParams{TenantID: f.tenantID, OrgID: f.orgID, FolderID: childFolder}

		// userB hidden by the restricted ANCESTOR folder (cascade).
		if mustFolderVisible(t, repo, withFolderUser(base, f.userB, false)) {
			t.Fatal("a child folder under a restricted parent must be hidden from a non-granted member")
		}

		// Grant read on the restricted parent -> cascades to the child.
		if _, err := repo.CreateGrant(context.Background(), CreateGrantParams{
			TenantID: f.tenantID, OrgID: f.orgID,
			ResourceType: "folder", ResourceID: parentFolder,
			PrincipalType: "user", PrincipalID: f.userB, Permission: "read",
		}); err != nil {
			t.Fatalf("CreateGrant (parent folder read): %v", err)
		}
		if !mustFolderVisible(t, repo, withFolderUser(base, f.userB, false)) {
			t.Fatal("a read grant on the restricted parent folder should make the child folder visible")
		}
	})
}

// TestFolderVisibility_CycleTerminates seeds a folder cycle (A.parent_id=B and
// B.parent_id=A) and asserts the cycle-protected recursive CTE in
// IsFolderVisible TERMINATES: the call must RETURN (not hang) within a 5s
// deadline. A regression that drops the path-accumulator/depth-cap cycle guard
// would loop forever and trip the context timeout.
func TestFolderVisibility_CycleTerminates(t *testing.T) {
	withTx(t, func(tx pgx.Tx) {
		f := seedBase(t, tx)

		// Insert both folders with NULL parents first (parent_id has an FK to
		// folders(id)), then wire the mutual cycle with UPDATEs.
		folderA := queryID(t, tx,
			`INSERT INTO folders (tenant_id, org_id, name, created_by, is_restricted)
			 VALUES ($1, $2, 'Folder A', $3, false) RETURNING id`, f.tenantID, f.orgID, f.userA)
		folderB := queryID(t, tx,
			`INSERT INTO folders (tenant_id, org_id, name, created_by, is_restricted)
			 VALUES ($1, $2, 'Folder B', $3, false) RETURNING id`, f.tenantID, f.orgID, f.userA)
		exec(t, tx, `UPDATE folders SET parent_id = $1 WHERE id = $2`, folderB, folderA)
		exec(t, tx, `UPDATE folders SET parent_id = $1 WHERE id = $2`, folderA, folderB)

		repo := aclRepoForTx(tx)
		params := FolderVisibilityParams{
			TenantID: f.tenantID, OrgID: f.orgID,
			FolderID: folderA, UserID: f.userB,
		}

		// The cycle-protected CTE must terminate; bound the call by 5s so a
		// regression that loses the guard fails fast instead of hanging the suite.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if _, err := repo.IsFolderVisible(ctx, params); err != nil {
			t.Fatalf("IsFolderVisible did not terminate on a folder cycle within 5s: %v", err)
		}
	})
}
