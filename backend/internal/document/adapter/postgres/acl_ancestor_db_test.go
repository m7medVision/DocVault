//go:build integration

// Characterization tests for the restricted-ancestor-folder visibility path
// through the two queries whose folder_ancestors CTE is seeded from candidate
// folders rather than every folder in the org (ListVisibleDocuments and
// SearchDocumentChunks). They assert that an open document filed under a
// restricted ancestor folder stays hidden from a non-granted member, and that a
// read grant on the ancestor restores it — i.e. the seed narrowing did not
// change the visibility decision.
package postgres

import (
	"context"
	"testing"

	"github.com/docvault/backend/internal/repository"
	"github.com/jackc/pgx/v5"
)

func TestVisibility_ListVisibleDocumentsRestrictedAncestorFolder(t *testing.T) {
	withTx(t, func(tx pgx.Tx) {
		f := seedBase(t, tx)

		parent := queryID(t, tx,
			`INSERT INTO folders (tenant_id, org_id, name, created_by, is_restricted)
			 VALUES ($1, $2, 'Confidential', $3, true) RETURNING id`, f.tenantID, f.orgID, f.userA)
		child := queryID(t, tx,
			`INSERT INTO folders (tenant_id, org_id, parent_id, name, created_by, is_restricted)
			 VALUES ($1, $2, $3, 'Sub', $4, false) RETURNING id`, f.tenantID, f.orgID, parent, f.userA)
		// An OPEN document, hidden only by the restricted ancestor folder.
		doc := queryID(t, tx,
			`INSERT INTO documents (tenant_id, org_id, folder_id, owner_id, title, is_restricted)
			 VALUES ($1, $2, $3, $4, 'Buried Doc', false) RETURNING id`, f.tenantID, f.orgID, child, f.userA)

		repo := aclRepoForTx(tx)
		listFor := func(userID string) map[string]bool {
			docs, _, err := repo.ListVisibleDocuments(context.Background(), repository.ListVisibleParams{
				TenantID: f.tenantID, OrgID: f.orgID, UserID: userID, Limit: 50,
			})
			if err != nil {
				t.Fatalf("ListVisibleDocuments(%s): %v", userID, err)
			}
			seen := map[string]bool{}
			for _, d := range docs {
				seen[d.ID] = true
			}
			return seen
		}

		// Non-granted member: hidden by the restricted ancestor.
		if listFor(f.userB)[doc] {
			t.Fatal("doc under a restricted ancestor folder must be excluded from a non-granted member's listing")
		}

		// Read grant on the ancestor (parent) folder restores visibility.
		if _, err := repo.CreateGrant(context.Background(), repository.CreateGrantParams{
			TenantID: f.tenantID, OrgID: f.orgID,
			ResourceType: "folder", ResourceID: parent,
			PrincipalType: "user", PrincipalID: f.userB, Permission: "read",
		}); err != nil {
			t.Fatalf("CreateGrant(folder read): %v", err)
		}
		if !listFor(f.userB)[doc] {
			t.Fatal("a read grant on the restricted ancestor folder should surface the doc in the listing")
		}
	})
}

func TestVisibility_SearchRestrictedAncestorFolder(t *testing.T) {
	withTx(t, func(tx pgx.Tx) {
		f := seedBase(t, tx)

		parent := queryID(t, tx,
			`INSERT INTO folders (tenant_id, org_id, name, created_by, is_restricted)
			 VALUES ($1, $2, 'Confidential', $3, true) RETURNING id`, f.tenantID, f.orgID, f.userA)
		child := queryID(t, tx,
			`INSERT INTO folders (tenant_id, org_id, parent_id, name, created_by, is_restricted)
			 VALUES ($1, $2, $3, 'Sub', $4, false) RETURNING id`, f.tenantID, f.orgID, parent, f.userA)
		doc := queryID(t, tx,
			`INSERT INTO documents (tenant_id, org_id, folder_id, owner_id, title, status, is_restricted)
			 VALUES ($1, $2, $3, $4, 'Buried Report', 'processed', false) RETURNING id`,
			f.tenantID, f.orgID, child, f.userA)
		versionID := queryID(t, tx,
			`INSERT INTO document_versions (document_id, version_number, storage_key, mime_type, file_size)
			 VALUES ($1, 1, 'key/1', 'application/pdf', 1024) RETURNING id`, doc)
		pageID := queryID(t, tx,
			`INSERT INTO document_pages (document_id, version_id, page_number, ocr_text)
			 VALUES ($1, $2, 1, 'buried contents') RETURNING id`, doc, versionID)
		embedding := embeddingLiteral(0.1)
		exec(t, tx,
			`INSERT INTO extracted_text_chunks (document_id, page_id, chunk_index, chunk_text, embedding)
			 VALUES ($1, $2, 0, 'buried contents about the merger', $3::vector)`, doc, pageID, embedding)

		search := searchRepoForTx(tx)
		req := repository.SearchRequest{
			Query:       "buried",
			QueryVector: embedding,
			TenantID:    f.tenantID,
			OrgID:       f.orgID,
			UserID:      f.userB,
			Limit:       20,
		}
		hasDoc := func(r repository.SearchRequest) bool {
			res, err := search.Search(context.Background(), r)
			if err != nil {
				t.Fatalf("Search: %v", err)
			}
			for _, c := range res.Chunks {
				if c.DocumentID == doc {
					return true
				}
			}
			return false
		}

		// Non-granted member: the restricted ancestor hides the chunk.
		if hasDoc(req) {
			t.Fatal("chunk under a restricted ancestor folder leaked to a non-granted member through search")
		}

		// is_admin bypass returns it.
		adminReq := req
		adminReq.IsAdmin = true
		if !hasDoc(adminReq) {
			t.Fatal("admin search should retrieve the chunk under a restricted ancestor folder")
		}

		// A read grant on the ancestor folder restores retrieval for the member.
		if _, err := aclRepoForTx(tx).CreateGrant(context.Background(), repository.CreateGrantParams{
			TenantID: f.tenantID, OrgID: f.orgID,
			ResourceType: "folder", ResourceID: parent,
			PrincipalType: "user", PrincipalID: f.userB, Permission: "read",
		}); err != nil {
			t.Fatalf("CreateGrant(folder read): %v", err)
		}
		if !hasDoc(req) {
			t.Fatal("a read grant on the restricted ancestor folder should surface the chunk in search")
		}
	})
}
