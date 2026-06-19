package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	sqldb "github.com/docvault/backend/internal/db"
	"github.com/docvault/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type searchRepository struct {
	queries sqldb.Querier
}

func NewSearchRepository(db *pgxpool.Pool) repository.SearchRepository {
	return &searchRepository{queries: sqldb.New(db)}
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

	rows, err := r.queries.SearchDocumentChunks(ctx, params)
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

func (r *searchRepository) DeleteChunksByDocument(ctx context.Context, docID string) error {
	return nil
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
