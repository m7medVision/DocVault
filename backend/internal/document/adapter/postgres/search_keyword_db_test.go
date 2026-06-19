//go:build integration

// Regression test for the proper-noun / keyword retrieval failure (the "teepee"
// bug). Before the fix, a keyword like "teepee" returned nothing because the
// lexical arm was a whole-phrase substring test and the semantic arm zeroed any
// cosine below 0.40. Here the query embedding is deliberately orthogonal to the
// chunk (cosine 0 -> semantic score 0), so only the FTS/trigram lexical arm can
// surface the chunk; an unrelated chunk with the same zero-cosine embedding and
// no lexical match must be filtered out.
package postgres

import (
	"context"
	"strings"
	"testing"

	"github.com/docvault/backend/internal/repository"
	"github.com/jackc/pgx/v5"
)

// oneHotEmbedding builds a 1024-dim unit vector with 1.0 at position pos and 0
// elsewhere. Two one-hot embeddings at different positions are orthogonal
// (cosine similarity 0), which lets a test isolate the lexical retrieval arm
// from the semantic one.
func oneHotEmbedding(pos int) string {
	parts := make([]string, 1024)
	for i := range parts {
		if i == pos {
			parts[i] = "1"
		} else {
			parts[i] = "0"
		}
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func TestSearch_KeywordRetrievalRescuesZeroCosusMatch(t *testing.T) {
	withTx(t, func(tx pgx.Tx) {
		f := seedBase(t, tx)

		// A document about a rare proper noun ("teepee").
		doc := queryID(t, tx,
			`INSERT INTO documents (tenant_id, org_id, owner_id, title, status, is_restricted)
			 VALUES ($1, $2, $3, 'Invoice TEEPEE-20260609', 'processed', false) RETURNING id`,
			f.tenantID, f.orgID, f.userA)
		versionID := queryID(t, tx,
			`INSERT INTO document_versions (document_id, version_number, storage_key, mime_type, file_size)
			 VALUES ($1, 1, 'k', 'application/pdf', 1) RETURNING id`, doc)
		pageID := queryID(t, tx,
			`INSERT INTO document_pages (document_id, version_id, page_number, ocr_text)
			 VALUES ($1, $2, 1, 'teepee invoice') RETURNING id`, doc, versionID)
		// Chunk embedding orthogonal to the query (dim 0) -> cosine 0.
		exec(t, tx,
			`INSERT INTO extracted_text_chunks (document_id, page_id, chunk_index, chunk_text, embedding)
			 VALUES ($1, $2, 0, 'Invoice TEEPEE-20260609 for OVH hosting', $3::vector)`,
			doc, pageID, oneHotEmbedding(0))

		// Unrelated document: same orthogonal embedding (cosine 0) and no lexical
		// match, so it must be filtered out -> proves the result is not noise.
		doc2 := queryID(t, tx,
			`INSERT INTO documents (tenant_id, org_id, owner_id, title, status, is_restricted)
			 VALUES ($1, $2, $3, 'Unrelated Memo', 'processed', false) RETURNING id`,
			f.tenantID, f.orgID, f.userA)
		v2 := queryID(t, tx,
			`INSERT INTO document_versions (document_id, version_number, storage_key, mime_type, file_size)
			 VALUES ($1, 1, 'k2', 'application/pdf', 1) RETURNING id`, doc2)
		p2 := queryID(t, tx,
			`INSERT INTO document_pages (document_id, version_id, page_number, ocr_text)
			 VALUES ($1, $2, 1, 'nothing relevant') RETURNING id`, doc2, v2)
		exec(t, tx,
			`INSERT INTO extracted_text_chunks (document_id, page_id, chunk_index, chunk_text, embedding)
			 VALUES ($1, $2, 0, 'completely unrelated weather report content', $3::vector)`,
			doc2, p2, oneHotEmbedding(0))

		search := searchRepoForTx(tx)
		// Query text "teepee" but an embedding orthogonal to the chunks (dim 1) ->
		// cosine 0, semantic score 0. Only the lexical arm can match.
		res, err := search.Search(context.Background(), repository.SearchRequest{
			Query:       "teepee",
			QueryVector: oneHotEmbedding(1),
			TenantID:    f.tenantID,
			OrgID:       f.orgID,
			UserID:      f.userA, // owner of both docs -> visible
			Limit:       20,
			MinScore:    0.01, // production floor
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}

		foundTeepee := false
		for _, c := range res.Chunks {
			if c.DocumentID == doc {
				foundTeepee = true
			}
			if c.DocumentID == doc2 {
				t.Fatal("unrelated chunk (no lexical match, cosine 0) leaked into keyword results")
			}
		}
		if !foundTeepee {
			t.Fatal("keyword query 'teepee' must retrieve the chunk via the lexical arm even when cosine is 0")
		}
	})
}
