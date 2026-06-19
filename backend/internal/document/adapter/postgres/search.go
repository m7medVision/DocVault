package postgres

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	sqldb "github.com/docvault/backend/internal/db"
	"github.com/docvault/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type searchRepository struct {
	pool     *pgxpool.Pool
	queries  sqldb.Querier
	efSearch int
}

// NewSearchRepository builds the search repository. efSearch, when > 0, sets
// hnsw.ef_search (via SET LOCAL) for each retrieval so recall at small K can be
// tuned without a redeploy; 0 leaves the Postgres default (40) in place.
func NewSearchRepository(db *pgxpool.Pool, efSearch int) repository.SearchRepository {
	return &searchRepository{pool: db, queries: sqldb.New(db), efSearch: efSearch}
}

func searchDocumentChunksParams(req repository.SearchRequest) (sqldb.SearchDocumentChunksParams, error) {
	if req.QueryVector == "" {
		return sqldb.SearchDocumentChunksParams{}, fmt.Errorf("query vector is required")
	}
	if strings.TrimSpace(req.Query) == "" {
		return sqldb.SearchDocumentChunksParams{}, fmt.Errorf("query text is required")
	}
	// Defense in depth: visibility filtering depends on a user identity, so
	// search must never run without one.
	if req.UserID == "" {
		return sqldb.SearchDocumentChunksParams{}, fmt.Errorf("user id is required")
	}

	limit := req.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}

	groupIDs := req.GroupIDs
	if groupIDs == nil {
		groupIDs = []string{}
	}

	return sqldb.SearchDocumentChunksParams{
		LimitCount:  int32(limit),
		QueryVector: req.QueryVector,
		QueryText:   strings.TrimSpace(req.Query),
		TenantID:    req.TenantID,
		OrgID:       req.OrgID,
		DocumentID:  optionalString(req.DocumentID),
		DocType:     optionalString(req.DocType),
		Language:    optionalString(req.Language),
		Status:      optionalString(req.Status),
		FolderID:    optionalString(req.FolderID),
		StartDate:   optionalTimestamptz(req.StartDate),
		EndDate:     optionalTimestamptz(req.EndDate),
		Tags:        tags,
		IsAdmin:     req.IsAdmin,
		UserID:      req.UserID,
		GroupIds:    groupIDs,
	}, nil
}

func (r *searchRepository) Search(ctx context.Context, req repository.SearchRequest) (*repository.SearchResult, error) {
	params, err := searchDocumentChunksParams(req)
	if err != nil {
		return nil, err
	}

	rows, err := r.searchChunks(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	chunks := make([]repository.ChunkMatch, 0, len(rows))
	for _, row := range rows {
		if req.MinScore > 0 && row.Score <= req.MinScore {
			continue
		}
		chunks = append(chunks, repository.ChunkMatch{
			DocumentID:    row.DocumentID,
			DocumentTitle: row.Title,
			DocType:       string(row.DocType),
			ChunkID:       row.ChunkID,
			ChunkText:     row.ChunkText,
			PageNumber:    int(row.PageNumber),
			Language:      row.Language,
			IsTranslation: row.IsTranslation,
			Score:         row.Score,
		})
	}

	return &repository.SearchResult{
		Chunks:     chunks,
		TotalCount: len(chunks),
	}, nil
}

func (r *searchRepository) IndexChunk(ctx context.Context, chunk repository.DocumentChunk) error {
	return nil
}

// searchChunks runs the retrieval query, optionally scoping hnsw.ef_search via
// SET LOCAL. SET LOCAL only applies inside a transaction, so when efSearch is
// configured the query runs in a short tx that rolls the GUC back on commit —
// it never leaks to other queries on the pooled connection. With efSearch == 0
// the query runs directly as before.
func (r *searchRepository) searchChunks(ctx context.Context, params sqldb.SearchDocumentChunksParams) ([]sqldb.SearchDocumentChunksRow, error) {
	if r.efSearch <= 0 {
		return r.queries.SearchDocumentChunks(ctx, params)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin ef_search tx: %w", err)
	}
	defer tx.Rollback(ctx) // no-op after a successful commit

	if _, err := tx.Exec(ctx, "SELECT set_config('hnsw.ef_search', $1, true)", strconv.Itoa(r.efSearch)); err != nil {
		return nil, fmt.Errorf("set hnsw.ef_search: %w", err)
	}

	rows, err := sqldb.New(tx).SearchDocumentChunks(ctx, params)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit ef_search tx: %w", err)
	}
	return rows, nil
}

func (r *searchRepository) DeleteChunksByDocument(ctx context.Context, docID string) error {
	return nil
}

// FetchDocumentsMetadata returns the extracted (or user-corrected) facts for
// the given documents, grouped by document id. A nil/empty value is skipped so
// only concrete facts reach the chat context. Tenant-scoped via the query.
func (r *searchRepository) FetchDocumentsMetadata(ctx context.Context, tenantID string, docIDs []string) (map[string][]repository.DocFact, error) {
	if len(docIDs) == 0 {
		return map[string][]repository.DocFact{}, nil
	}
	rows, err := r.queries.GetDocumentsMetadata(ctx, sqldb.GetDocumentsMetadataParams{
		TenantID:    tenantID,
		DocumentIds: docIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("fetch document metadata: %w", err)
	}
	out := make(map[string][]repository.DocFact, len(rows))
	for _, row := range rows {
		if row.Value == nil {
			continue
		}
		out[row.DocumentID] = append(out[row.DocumentID], repository.DocFact{
			DocumentID: row.DocumentID,
			Key:        row.Key,
			Value:      *row.Value,
		})
	}
	return out, nil
}

func optionalString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func optionalTimestamptz(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *value, Valid: true}
}
