package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

const semanticScoreFloor = 0.4

type DocumentChunk struct {
	ID         string
	DocumentID string
	ChunkText  string
	PageNumber int
	Embedding  []float32
	TenantID   string
}

type ChunkMatch struct {
	DocumentID    string
	DocumentTitle string
	DocType       string
	ChunkID       string
	ChunkText     string
	PageNumber    int
	Language      string
	IsTranslation bool
	Score         float64
}

type SearchRequest struct {
	Query             string
	TenantID          string
	UserID            string
	OrgID             string
	FilterMeta        map[string]interface{}
	Limit             int
	Cursor            string
	QueryVector       string
	FilterWhereClause string
	FilterParams      []interface{}
	MinScore          float64
}

type SearchResult struct {
	Chunks     []ChunkMatch
	NextCursor string
	TotalCount int
}

type SearchRepository interface {
	Search(ctx context.Context, req SearchRequest) (*SearchResult, error)
	IndexChunk(ctx context.Context, chunk DocumentChunk) error
	DeleteChunksByDocument(ctx context.Context, docID string) error
}

type searchRepository struct {
	db *pgxpool.Pool
}

func NewSearchRepository(db *pgxpool.Pool) SearchRepository {
	return &searchRepository{db: db}
}

func buildSearchQuery(req SearchRequest) (string, []interface{}, error) {
	if req.QueryVector == "" {
		return "", nil, fmt.Errorf("query vector is required")
	}
	if strings.TrimSpace(req.Query) == "" {
		return "", nil, fmt.Errorf("query text is required")
	}

	limit := req.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	queryTextPlaceholder := fmt.Sprintf("$%d", len(req.FilterParams)+2)
	chunkContainsExpr := fmt.Sprintf("POSITION(LOWER(%s) IN LOWER(c.chunk_text)) > 0", queryTextPlaceholder)
	titleContainsExpr := fmt.Sprintf("POSITION(LOWER(%s) IN LOWER(d.title)) > 0", queryTextPlaceholder)
	rawSemanticScoreExpr := "GREATEST(0.0, 1 - (c.embedding <=> $1::vector))"
	semanticScoreExpr := fmt.Sprintf(
		"LEAST(1.0, GREATEST(0.0, ((%s) - %.2f) / %.2f))",
		rawSemanticScoreExpr,
		semanticScoreFloor,
		1.0-semanticScoreFloor,
	)
	lexicalScoreExpr := fmt.Sprintf(`CASE
			WHEN LOWER(BTRIM(d.title)) = LOWER(BTRIM(%s)) THEN 1.0
			WHEN LOWER(BTRIM(c.chunk_text)) = LOWER(BTRIM(%s)) THEN 0.99
			WHEN %s THEN 0.97
			WHEN %s THEN 0.95
			ELSE 0.0
		END`, queryTextPlaceholder, queryTextPlaceholder, chunkContainsExpr, titleContainsExpr)
	hybridScoreExpr := fmt.Sprintf("GREATEST(%s, %s)", semanticScoreExpr, lexicalScoreExpr)

	baseQuery := fmt.Sprintf(`
		SELECT 
			c.document_id,
			d.title,
			d.doc_type,
			c.id as chunk_id,
			c.chunk_text,
			COALESCE(p.page_number, 0) as page_number,
			COALESCE(d.language, '') as language,
			FALSE as is_translation,
			%s as score,
			c.embedding <=> $1::vector as distance
		FROM extracted_text_chunks c
		JOIN documents d ON c.document_id = d.id
		LEFT JOIN document_pages p ON c.page_id = p.id
		WHERE c.embedding IS NOT NULL`, hybridScoreExpr)

	if req.FilterWhereClause != "" {
		baseQuery += " AND " + req.FilterWhereClause
	}

	lexicalFilter := fmt.Sprintf("(%s OR %s)", chunkContainsExpr, titleContainsExpr)

	query := fmt.Sprintf(`
		WITH vector_matches AS (
			%s
			ORDER BY c.embedding <=> $1::vector
			LIMIT %d
		),
		lexical_matches AS (
			%s
			AND %s
			ORDER BY score DESC, c.embedding <=> $1::vector
			LIMIT %d
		),
		candidate_matches AS (
			SELECT * FROM vector_matches
			UNION ALL
			SELECT * FROM lexical_matches
		),
		deduped_matches AS (
			SELECT DISTINCT ON (chunk_id)
				document_id,
				title,
				doc_type,
				chunk_id,
				chunk_text,
				page_number,
				language,
				is_translation,
				score,
				distance
			FROM candidate_matches
			ORDER BY chunk_id, score DESC, distance ASC
		)
		SELECT 
			document_id,
			title,
			doc_type,
			chunk_id,
			chunk_text,
			page_number,
			language,
			is_translation,
			score
		FROM deduped_matches`, baseQuery, limit, baseQuery, lexicalFilter, limit)

	args := make([]interface{}, 0, len(req.FilterParams)+3)
	args = append(args, req.QueryVector)
	args = append(args, req.FilterParams...)
	args = append(args, strings.TrimSpace(req.Query))

	if req.MinScore > 0 {
		thresholdPlaceholder := fmt.Sprintf("$%d", len(args)+1)
		query += " WHERE score > " + thresholdPlaceholder
		args = append(args, req.MinScore)
	}

	query += fmt.Sprintf(" ORDER BY score DESC LIMIT %d", limit)

	return query, args, nil
}

func (r *searchRepository) Search(ctx context.Context, req SearchRequest) (*SearchResult, error) {
	query, args, err := buildSearchQuery(req)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	defer rows.Close()

	var chunks []ChunkMatch
	for rows.Next() {
		var c ChunkMatch
		if err := rows.Scan(
			&c.DocumentID, &c.DocumentTitle, &c.DocType,
			&c.ChunkID, &c.ChunkText, &c.PageNumber,
			&c.Language, &c.IsTranslation, &c.Score,
		); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		chunks = append(chunks, c)
	}

	return &SearchResult{
		Chunks:     chunks,
		TotalCount: len(chunks),
	}, nil
}

func (r *searchRepository) IndexChunk(ctx context.Context, chunk DocumentChunk) error {
	return nil
}

func (r *searchRepository) DeleteChunksByDocument(ctx context.Context, docID string) error {
	return nil
}
