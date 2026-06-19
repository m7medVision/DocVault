//go:build integration

// Package repository DB-backed integration tests for per-document/folder
// visibility. These are excluded from the default `go test ./...` run by the
// `integration` build tag and require a live Postgres reachable at DATABASE_URL
// (migrated to at least 013). Every test runs inside a transaction that is
// always rolled back, so the target database is left untouched.
package postgres

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	sqldb "github.com/docvault/backend/internal/db"
	"github.com/docvault/backend/internal/migrate"
	"github.com/docvault/backend/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultDatabaseURL = "postgresql://docvault:docvault_dev@localhost:5432/docvault"

func databaseURL() string {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	return defaultDatabaseURL
}

// embeddingLiteral builds a 1024-dimension pgvector literal filled with v.
func embeddingLiteral(v float64) string {
	parts := make([]string, 1024)
	for i := range parts {
		parts[i] = fmt.Sprintf("%g", v)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// visibilityFixture holds the identifiers seeded for a single test case.
type visibilityFixture struct {
	tenantID   string
	orgID      string
	userA      string // owner of the restricted document
	userB      string // other member, granted nothing by default
	folderID   string // folder the document lives in (may be empty)
	documentID string
}

// exec runs a statement inside the test transaction, failing the test on error.
func exec(t *testing.T, tx pgx.Tx, sql string, args ...interface{}) {
	t.Helper()
	if _, err := tx.Exec(context.Background(), sql, args...); err != nil {
		t.Fatalf("exec failed: %v\nSQL: %s", err, sql)
	}
}

// queryID runs a RETURNING/SELECT statement expected to yield a single uuid.
func queryID(t *testing.T, tx pgx.Tx, sql string, args ...interface{}) string {
	t.Helper()
	var id string
	if err := tx.QueryRow(context.Background(), sql, args...).Scan(&id); err != nil {
		t.Fatalf("query id failed: %v\nSQL: %s", err, sql)
	}
	return id
}

// seedBase creates a tenant, org, and two users and returns the partially
// populated fixture; callers add folders/documents/grants on top.
func seedBase(t *testing.T, tx pgx.Tx) visibilityFixture {
	t.Helper()

	var f visibilityFixture
	f.tenantID = queryID(t, tx,
		`INSERT INTO tenants (name, plan) VALUES ('acl-test', 'business') RETURNING id`)
	f.orgID = queryID(t, tx,
		`INSERT INTO organizations (tenant_id, name) VALUES ($1, 'acl-org') RETURNING id`, f.tenantID)
	f.userA = queryID(t, tx,
		`INSERT INTO users (tenant_id, email, display_name) VALUES ($1, 'a@acl.test', 'User A') RETURNING id`, f.tenantID)
	f.userB = queryID(t, tx,
		`INSERT INTO users (tenant_id, email, display_name) VALUES ($1, 'b@acl.test', 'User B') RETURNING id`, f.tenantID)
	return f
}

// withTx opens a transaction, runs fn, and always rolls back so the DB stays
// clean regardless of assertions.
func withTx(t *testing.T, fn func(tx pgx.Tx)) {
	t.Helper()
	ctx := context.Background()

	if err := migrate.Run(ctx, databaseURL()); err != nil {
		t.Fatalf("migrate.Run: %v", err)
	}

	pool, err := pgxpool.New(ctx, databaseURL())
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

// aclRepoForTx wires the repository implementation to the test transaction so
// the production code path (sqldb.New + queries) is exercised end to end.
func aclRepoForTx(tx pgx.Tx) *aclRepository {
	return &aclRepository{queries: sqldb.New(tx)}
}

func searchRepoForTx(tx pgx.Tx) *searchRepository {
	return &searchRepository{queries: sqldb.New(tx)}
}

func mustVisible(t *testing.T, repo *aclRepository, params repository.VisibilityParams) bool {
	t.Helper()
	visible, err := repo.IsDocumentVisible(context.Background(), params)
	if err != nil {
		t.Fatalf("IsDocumentVisible error: %v", err)
	}
	return visible
}

func mustWritable(t *testing.T, repo *aclRepository, params repository.VisibilityParams) bool {
	t.Helper()
	writable, err := repo.IsDocumentWritable(context.Background(), params)
	if err != nil {
		t.Fatalf("IsDocumentWritable error: %v", err)
	}
	return writable
}

// TestVisibility_RestrictedDocumentGrantAndAdminBypass covers the core rule:
// a restricted doc is hidden from a non-granted member, visible to its owner,
// visible to a user after a read grant, and always visible when is_admin=true.
func TestVisibility_RestrictedDocumentGrantAndAdminBypass(t *testing.T) {
	withTx(t, func(tx pgx.Tx) {
		f := seedBase(t, tx)
		f.documentID = queryID(t, tx,
			`INSERT INTO documents (tenant_id, org_id, owner_id, title, is_restricted)
			 VALUES ($1, $2, $3, 'Restricted Doc', true) RETURNING id`,
			f.tenantID, f.orgID, f.userA)

		repo := aclRepoForTx(tx)
		base := repository.VisibilityParams{TenantID: f.tenantID, OrgID: f.orgID, DocumentID: f.documentID}

		// userB: not owner, not granted, not admin -> NOT visible.
		if mustVisible(t, repo, withUser(base, f.userB, false)) {
			t.Fatal("userB should not see a restricted doc they do not own and were not granted")
		}

		// owner userA -> visible.
		if !mustVisible(t, repo, withUser(base, f.userA, false)) {
			t.Fatal("owner userA should see their own restricted doc")
		}

		// is_admin bypass for userB -> visible.
		adminParams := withUser(base, f.userB, true)
		if !mustVisible(t, repo, adminParams) {
			t.Fatal("is_admin=true should bypass restriction")
		}

		// repository.Grant read to userB on the document -> visible.
		grantID, err := repo.CreateGrant(context.Background(), repository.CreateGrantParams{
			TenantID: f.tenantID, OrgID: f.orgID,
			ResourceType: "document", ResourceID: f.documentID,
			PrincipalType: "user", PrincipalID: f.userB, Permission: "read",
		})
		if err != nil {
			t.Fatalf("CreateGrant: %v", err)
		}
		if grantID == "" {
			t.Fatal("CreateGrant returned empty id for a new grant")
		}
		if !mustVisible(t, repo, withUser(base, f.userB, false)) {
			t.Fatal("userB should see the doc after a read grant")
		}
	})
}

// TestVisibility_MissingDocumentIsNotVisible asserts that an unknown document
// maps to (false, nil): missing is indistinguishable from not-visible.
func TestVisibility_MissingDocumentIsNotVisible(t *testing.T) {
	withTx(t, func(tx pgx.Tx) {
		f := seedBase(t, tx)
		repo := aclRepoForTx(tx)
		visible, err := repo.IsDocumentVisible(context.Background(), repository.VisibilityParams{
			TenantID: f.tenantID, OrgID: f.orgID,
			DocumentID: "00000000-0000-0000-0000-000000000000",
			UserID:     f.userB,
		})
		if err != nil {
			t.Fatalf("IsDocumentVisible error: %v", err)
		}
		if visible {
			t.Fatal("a missing document must not be visible")
		}
	})
}

// TestVisibility_FolderRestrictionCascade asserts a doc under a restricted
// ancestor folder is hidden, and a read grant on the ancestor folder restores
// visibility (cascade through the recursive folder chain).
func TestVisibility_FolderRestrictionCascade(t *testing.T) {
	withTx(t, func(tx pgx.Tx) {
		f := seedBase(t, tx)

		// Two-level folder tree: parent (restricted) -> child (open).
		parentFolder := queryID(t, tx,
			`INSERT INTO folders (tenant_id, org_id, name, is_restricted)
			 VALUES ($1, $2, 'Restricted Parent', true) RETURNING id`, f.tenantID, f.orgID)
		childFolder := queryID(t, tx,
			`INSERT INTO folders (tenant_id, org_id, parent_id, name, is_restricted)
			 VALUES ($1, $2, $3, 'Open Child', false) RETURNING id`, f.tenantID, f.orgID, parentFolder)

		// Document itself is not restricted, but lives under the restricted parent.
		f.documentID = queryID(t, tx,
			`INSERT INTO documents (tenant_id, org_id, folder_id, owner_id, title, is_restricted)
			 VALUES ($1, $2, $3, $4, 'Doc Under Restricted Folder', false) RETURNING id`,
			f.tenantID, f.orgID, childFolder, f.userA)

		repo := aclRepoForTx(tx)
		base := repository.VisibilityParams{TenantID: f.tenantID, OrgID: f.orgID, DocumentID: f.documentID}

		// userB hidden by the restricted ancestor folder.
		if mustVisible(t, repo, withUser(base, f.userB, false)) {
			t.Fatal("doc under a restricted ancestor folder must be hidden from a non-granted member")
		}

		// repository.Grant read on the ancestor (parent) folder -> cascades to the doc.
		if _, err := repo.CreateGrant(context.Background(), repository.CreateGrantParams{
			TenantID: f.tenantID, OrgID: f.orgID,
			ResourceType: "folder", ResourceID: parentFolder,
			PrincipalType: "user", PrincipalID: f.userB, Permission: "read",
		}); err != nil {
			t.Fatalf("CreateGrant (folder): %v", err)
		}
		if !mustVisible(t, repo, withUser(base, f.userB, false)) {
			t.Fatal("a read grant on the restricted ancestor folder should make the doc visible")
		}
	})
}

// TestVisibility_GroupGrant asserts a grant to a group is honoured for a member
// of that group (group ids resolved via ListUserGroupIDs).
func TestVisibility_GroupGrant(t *testing.T) {
	withTx(t, func(tx pgx.Tx) {
		f := seedBase(t, tx)
		f.documentID = queryID(t, tx,
			`INSERT INTO documents (tenant_id, org_id, owner_id, title, is_restricted)
			 VALUES ($1, $2, $3, 'repository.Group Restricted Doc', true) RETURNING id`,
			f.tenantID, f.orgID, f.userA)

		repo := aclRepoForTx(tx)

		// Create a group and add userB to it.
		group, err := repo.CreateGroup(context.Background(), repository.CreateGroupParams{
			TenantID: f.tenantID, OrgID: f.orgID, Name: "Legal",
		})
		if err != nil {
			t.Fatalf("CreateGroup: %v", err)
		}
		if err := repo.AddGroupMember(context.Background(), f.tenantID, f.orgID, group.ID, f.userB); err != nil {
			t.Fatalf("AddGroupMember: %v", err)
		}

		groupIDs, err := repo.ListUserGroupIDs(context.Background(), f.userB, f.orgID)
		if err != nil {
			t.Fatalf("ListUserGroupIDs: %v", err)
		}
		if len(groupIDs) != 1 || groupIDs[0] != group.ID {
			t.Fatalf("ListUserGroupIDs = %#v, want [%s]", groupIDs, group.ID)
		}

		base := repository.VisibilityParams{
			TenantID: f.tenantID, OrgID: f.orgID, DocumentID: f.documentID,
			UserID: f.userB, GroupIDs: groupIDs,
		}

		// Before any grant, the group member still cannot see it.
		if mustVisible(t, repo, base) {
			t.Fatal("group member should not see a restricted doc before a grant")
		}

		// repository.Grant read to the group -> the member can now see it.
		if _, err := repo.CreateGrant(context.Background(), repository.CreateGrantParams{
			TenantID: f.tenantID, OrgID: f.orgID,
			ResourceType: "document", ResourceID: f.documentID,
			PrincipalType: "group", PrincipalID: group.ID, Permission: "read",
		}); err != nil {
			t.Fatalf("CreateGrant (group): %v", err)
		}
		if !mustVisible(t, repo, base) {
			t.Fatal("group member should see the doc after a grant to the group")
		}
	})
}

// TestVisibility_FolderIDNullRestrictedDoc asserts the recursive folder chain
// degenerates correctly when folder_id IS NULL: a restricted root-level doc is
// hidden from a non-granted member but visible to its owner (no SQL error from
// the empty chain).
func TestVisibility_FolderIDNullRestrictedDoc(t *testing.T) {
	withTx(t, func(tx pgx.Tx) {
		f := seedBase(t, tx)
		f.documentID = queryID(t, tx,
			`INSERT INTO documents (tenant_id, org_id, folder_id, owner_id, title, is_restricted)
			 VALUES ($1, $2, NULL, $3, 'Root Restricted Doc', true) RETURNING id`,
			f.tenantID, f.orgID, f.userA)

		repo := aclRepoForTx(tx)
		base := repository.VisibilityParams{TenantID: f.tenantID, OrgID: f.orgID, DocumentID: f.documentID}

		if mustVisible(t, repo, withUser(base, f.userB, false)) {
			t.Fatal("restricted root-level (folder_id NULL) doc must be hidden from a non-granted member")
		}
		if !mustVisible(t, repo, withUser(base, f.userA, false)) {
			t.Fatal("owner should see their restricted root-level doc")
		}
	})
}

// TestVisibility_ListVisibleDocumentsFiltersRestricted asserts the list path
// hides restricted docs from a non-granted member while including open docs and
// owner-owned docs.
func TestVisibility_ListVisibleDocumentsFiltersRestricted(t *testing.T) {
	withTx(t, func(tx pgx.Tx) {
		f := seedBase(t, tx)

		openDoc := queryID(t, tx,
			`INSERT INTO documents (tenant_id, org_id, owner_id, title, is_restricted)
			 VALUES ($1, $2, $3, 'Open Doc', false) RETURNING id`, f.tenantID, f.orgID, f.userA)
		restrictedDoc := queryID(t, tx,
			`INSERT INTO documents (tenant_id, org_id, owner_id, title, is_restricted)
			 VALUES ($1, $2, $3, 'Restricted Doc', true) RETURNING id`, f.tenantID, f.orgID, f.userA)

		repo := aclRepoForTx(tx)
		docs, _, err := repo.ListVisibleDocuments(context.Background(), repository.ListVisibleParams{
			TenantID: f.tenantID, OrgID: f.orgID, UserID: f.userB, Limit: 50,
		})
		if err != nil {
			t.Fatalf("ListVisibleDocuments: %v", err)
		}

		seen := map[string]bool{}
		for _, d := range docs {
			seen[d.ID] = true
		}
		if !seen[openDoc] {
			t.Fatal("open doc should be visible to a member in the listing")
		}
		if seen[restrictedDoc] {
			t.Fatal("restricted doc should be excluded from a non-granted member's listing")
		}
	})
}

// TestVisibility_SearchFiltersRestrictedChunks seeds a chunk+embedding for a
// restricted document and asserts SearchDocumentChunks does not return that
// chunk for a non-granted user, but does once is_admin bypasses (proving the
// retrieval seam enforces visibility).
func TestVisibility_SearchFiltersRestrictedChunks(t *testing.T) {
	withTx(t, func(tx pgx.Tx) {
		f := seedBase(t, tx)
		f.documentID = queryID(t, tx,
			`INSERT INTO documents (tenant_id, org_id, owner_id, title, status, is_restricted)
			 VALUES ($1, $2, $3, 'Secret Report', 'processed', true) RETURNING id`,
			f.tenantID, f.orgID, f.userA)

		versionID := queryID(t, tx,
			`INSERT INTO document_versions (document_id, version_number, storage_key, mime_type, file_size)
			 VALUES ($1, 1, 'key/1', 'application/pdf', 1024) RETURNING id`, f.documentID)
		pageID := queryID(t, tx,
			`INSERT INTO document_pages (document_id, version_id, page_number, ocr_text)
			 VALUES ($1, $2, 1, 'secret contents') RETURNING id`, f.documentID, versionID)

		embedding := embeddingLiteral(0.1)
		exec(t, tx,
			`INSERT INTO extracted_text_chunks (document_id, page_id, chunk_index, chunk_text, embedding)
			 VALUES ($1, $2, 0, 'secret contents about the merger', $3::vector)`,
			f.documentID, pageID, embedding)

		search := searchRepoForTx(tx)
		req := repository.SearchRequest{
			Query:       "secret",
			QueryVector: embedding,
			TenantID:    f.tenantID,
			OrgID:       f.orgID,
			UserID:      f.userB,
			Limit:       20,
		}

		// Non-granted member: the restricted chunk must be filtered out.
		res, err := search.Search(context.Background(), req)
		if err != nil {
			t.Fatalf("Search (member): %v", err)
		}
		for _, c := range res.Chunks {
			if c.DocumentID == f.documentID {
				t.Fatal("restricted document's chunk leaked to a non-granted member through search")
			}
		}

		// is_admin bypass: the chunk is retrievable.
		adminReq := req
		adminReq.IsAdmin = true
		adminRes, err := search.Search(context.Background(), adminReq)
		if err != nil {
			t.Fatalf("Search (admin): %v", err)
		}
		found := false
		for _, c := range adminRes.Chunks {
			if c.DocumentID == f.documentID {
				found = true
			}
		}
		if !found {
			t.Fatal("admin search should retrieve the restricted document's chunk")
		}
	})
}

// TestWritability_ReadGrantDoesNotConferWrite covers write-permission
// enforcement: on a restricted document a read grant grants visibility but NOT
// writability; only a write (or delete) grant makes the doc writable. The owner
// and is_admin=true are always writable.
func TestWritability_ReadGrantDoesNotConferWrite(t *testing.T) {
	withTx(t, func(tx pgx.Tx) {
		f := seedBase(t, tx)
		f.documentID = queryID(t, tx,
			`INSERT INTO documents (tenant_id, org_id, owner_id, title, is_restricted)
			 VALUES ($1, $2, $3, 'Restricted Doc', true) RETURNING id`,
			f.tenantID, f.orgID, f.userA)

		repo := aclRepoForTx(tx)
		base := repository.VisibilityParams{TenantID: f.tenantID, OrgID: f.orgID, DocumentID: f.documentID}

		// repository.Grant userB a READ grant on the document.
		if _, err := repo.CreateGrant(context.Background(), repository.CreateGrantParams{
			TenantID: f.tenantID, OrgID: f.orgID,
			ResourceType: "document", ResourceID: f.documentID,
			PrincipalType: "user", PrincipalID: f.userB, Permission: "read",
		}); err != nil {
			t.Fatalf("CreateGrant (read): %v", err)
		}

		// A read grant confers visibility but must NOT confer writability.
		if !mustVisible(t, repo, withUser(base, f.userB, false)) {
			t.Fatal("userB should see the doc after a read grant")
		}
		if mustWritable(t, repo, withUser(base, f.userB, false)) {
			t.Fatal("a read grant must NOT confer write permission on a restricted doc")
		}

		// repository.Grant userB a WRITE grant on the document -> now writable.
		if _, err := repo.CreateGrant(context.Background(), repository.CreateGrantParams{
			TenantID: f.tenantID, OrgID: f.orgID,
			ResourceType: "document", ResourceID: f.documentID,
			PrincipalType: "user", PrincipalID: f.userB, Permission: "write",
		}); err != nil {
			t.Fatalf("CreateGrant (write): %v", err)
		}
		if !mustWritable(t, repo, withUser(base, f.userB, false)) {
			t.Fatal("userB should be able to write the doc after a write grant")
		}

		// Owner is always writable.
		if !mustWritable(t, repo, withUser(base, f.userA, false)) {
			t.Fatal("owner userA should be able to write their own restricted doc")
		}

		// is_admin=true bypasses restriction and is writable.
		if !mustWritable(t, repo, withUser(base, f.userB, true)) {
			t.Fatal("is_admin=true should be able to write a restricted doc")
		}
	})
}

// TestVisibility_FolderCycleTerminates seeds a folder cycle (A.parent_id=B and
// B.parent_id=A) with a document in A and asserts the cycle-protected recursive
// CTE terminates: IsDocumentVisible and IsDocumentWritable must RETURN (not
// hang) within a 5s deadline. A regression that drops the cycle guard would
// loop forever and trip the context timeout.
func TestVisibility_FolderCycleTerminates(t *testing.T) {
	withTx(t, func(tx pgx.Tx) {
		f := seedBase(t, tx)

		// Insert both folders with NULL parents first (parent_id has an FK to
		// folders(id)), then wire the mutual cycle with UPDATEs.
		folderA := queryID(t, tx,
			`INSERT INTO folders (tenant_id, org_id, name, is_restricted)
			 VALUES ($1, $2, 'Folder A', false) RETURNING id`, f.tenantID, f.orgID)
		folderB := queryID(t, tx,
			`INSERT INTO folders (tenant_id, org_id, name, is_restricted)
			 VALUES ($1, $2, 'Folder B', false) RETURNING id`, f.tenantID, f.orgID)
		exec(t, tx, `UPDATE folders SET parent_id = $1 WHERE id = $2`, folderB, folderA)
		exec(t, tx, `UPDATE folders SET parent_id = $1 WHERE id = $2`, folderA, folderB)

		// Document lives in folder A, which sits on the cycle.
		f.documentID = queryID(t, tx,
			`INSERT INTO documents (tenant_id, org_id, folder_id, owner_id, title, is_restricted)
			 VALUES ($1, $2, $3, $4, 'Doc In Cyclic Folder', false) RETURNING id`,
			f.tenantID, f.orgID, folderA, f.userA)

		repo := aclRepoForTx(tx)
		params := repository.VisibilityParams{
			TenantID: f.tenantID, OrgID: f.orgID,
			DocumentID: f.documentID, UserID: f.userB,
		}

		// The cycle-protected CTE must terminate; bound each call by 5s so a
		// regression that loses the guard fails fast instead of hanging the suite.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if _, err := repo.IsDocumentVisible(ctx, params); err != nil {
			t.Fatalf("IsDocumentVisible did not terminate on a folder cycle within 5s: %v", err)
		}
		if _, err := repo.IsDocumentWritable(ctx, params); err != nil {
			t.Fatalf("IsDocumentWritable did not terminate on a folder cycle within 5s: %v", err)
		}
	})
}

// withUser returns a copy of params with the user identity and admin flag set.
func withUser(params repository.VisibilityParams, userID string, isAdmin bool) repository.VisibilityParams {
	params.UserID = userID
	params.IsAdmin = isAdmin
	return params
}
