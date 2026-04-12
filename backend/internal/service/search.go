package service

import (
	"context"
	"fmt"
	"log/slog"
	"sort"

	"github.com/docvault/backend/internal/repository"
	"github.com/docvault/backend/internal/search"
)

const expectedEmbeddingDimensions = 1024
const minimumSearchScore = 0.01
const chunkCandidateMultiplier = 5
const maxChunkCandidates = 200
const reciprocalRankConstant = 60.0

type SearchService struct {
	embedder   search.Embedder
	searchRepo repository.SearchRepository
}

func NewSearchService(embedder search.Embedder, searchRepo repository.SearchRepository) *SearchService {
	return &SearchService{
		embedder:   embedder,
		searchRepo: searchRepo,
	}
}

type SearchInput struct {
	Query     string
	Limit     int
	DocType   string
	Language  string
	Status    string
	FolderID  string
	Tags      []string
	TenantID  string
	OrgID     string
	StartDate string
	EndDate   string
}

type SearchResult struct {
	DocumentID string  `json:"document_id"`
	File       string  `json:"file"`
	MaxScore   float64 `json:"max_score"`
}

type aggregatedSearchResult struct {
	SearchResult
	rankScore float64
}

type SearchOutput struct {
	Results    []SearchResult `json:"results"`
	Query      string         `json:"query"`
	Total      int            `json:"total"`
	Page       int            `json:"page,omitempty"`
	PageSize   int            `json:"page_size,omitempty"`
	HasMore    bool           `json:"has_more,omitempty"`
	NextCursor string         `json:"next_cursor,omitempty"`
}

type SearchResultWithMetadata struct {
	Results      []SearchResult `json:"results"`
	Query        string         `json:"query"`
	Total        int            `json:"total"`
	Page         int            `json:"page"`
	PageSize     int            `json:"page_size"`
	TotalPages   int            `json:"total_pages,omitempty"`
	HasMore      bool           `json:"has_more"`
	NextCursor   string         `json:"next_cursor,omitempty"`
	SearchTimeMs int64          `json:"search_time_ms,omitempty"`
	Filters      AppliedFilters `json:"filters,omitempty"`
}

type AppliedFilters struct {
	DocType   string   `json:"doc_type,omitempty"`
	Language  string   `json:"language,omitempty"`
	Status    string   `json:"status,omitempty"`
	FolderID  string   `json:"folder_id,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	StartDate string   `json:"start_date,omitempty"`
	EndDate   string   `json:"end_date,omitempty"`
}

type FilterSQL struct {
	WhereClause string
	Params      []interface{}
	ParamCount  int
}

func NewFilterSQL() *FilterSQL {
	return &FilterSQL{
		WhereClause: "",
		Params:      []interface{}{},
		ParamCount:  0,
	}
}

func (f *FilterSQL) AddCondition(condition string) {
	if f.WhereClause == "" {
		f.WhereClause = condition
	} else {
		f.WhereClause += " AND " + condition
	}
}

func (f *FilterSQL) addParam(param interface{}) string {
	f.ParamCount++
	f.Params = append(f.Params, param)
	return fmt.Sprintf("$%d", f.ParamCount+1)
}

func (f *FilterSQL) Build() (string, []interface{}) {
	return f.WhereClause, f.Params
}

func GenerateFilterSQL(input *SearchInput) *FilterSQL {
	f := NewFilterSQL()

	if input.TenantID != "" {
		f.AddCondition(fmt.Sprintf("d.tenant_id = %s", f.addParam(input.TenantID)))
	}
	if input.OrgID != "" {
		f.AddCondition(fmt.Sprintf("d.org_id = %s", f.addParam(input.OrgID)))
	}
	if input.DocType != "" {
		f.AddCondition(fmt.Sprintf("d.doc_type = %s", f.addParam(input.DocType)))
	}
	if input.Language != "" {
		f.AddCondition(fmt.Sprintf("d.language = %s", f.addParam(input.Language)))
	}
	if input.Status != "" {
		f.AddCondition(fmt.Sprintf("d.status = %s", f.addParam(input.Status)))
	}
	if input.FolderID != "" {
		f.AddCondition(fmt.Sprintf("d.folder_id = %s", f.addParam(input.FolderID)))
	}
	if input.StartDate != "" {
		f.AddCondition(fmt.Sprintf("d.created_at >= %s", f.addParam(input.StartDate)))
	}
	if input.EndDate != "" {
		f.AddCondition(fmt.Sprintf("d.created_at <= %s", f.addParam(input.EndDate)))
	}
	if len(input.Tags) > 0 {
		for _, tag := range input.Tags {
			f.AddCondition(fmt.Sprintf(
				"EXISTS (SELECT 1 FROM document_tags dt JOIN tags t ON t.id = dt.tag_id WHERE dt.document_id = d.id AND t.tenant_id = d.tenant_id AND t.name = %s)",
				f.addParam(tag),
			))
		}
	}

	return f
}

func BuildVectorSearchQuery(filterSQL *FilterSQL, limit int) (string, []interface{}) {
	baseQuery := `
		SELECT 
			c.document_id,
			d.title,
			d.doc_type,
			c.id as chunk_id,
			c.chunk_text,
			COALESCE(p.page_number, 0) as page_number,
			COALESCE(d.language, '') as language,
			FALSE as is_translation,
			(c.embedding <=> $1::vector) as distance
		FROM extracted_text_chunks c
		JOIN documents d ON c.document_id = d.id
		LEFT JOIN document_pages p ON c.page_id = p.id
	`

	if filterSQL.WhereClause != "" {
		return baseQuery + " WHERE " + filterSQL.WhereClause + fmt.Sprintf(" ORDER BY c.embedding <=> $1::vector LIMIT %d", limit), filterSQL.Params
	}

	return baseQuery + fmt.Sprintf(" ORDER BY c.embedding <=> $1::vector LIMIT %d", limit), []interface{}{}
}

func expandChunkCandidateLimit(limit int) int {
	expanded := limit * chunkCandidateMultiplier
	if expanded > maxChunkCandidates {
		return maxChunkCandidates
	}
	return expanded
}

func reciprocalRank(rank int) float64 {
	return 1.0 / (reciprocalRankConstant + float64(rank) + 1)
}

func (s *SearchService) Search(ctx context.Context, input *SearchInput) (*SearchOutput, error) {
	if input.Query == "" {
		return nil, fmt.Errorf("query is required")
	}

	limit := input.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	embedding, err := s.embedder.Embed(ctx, input.Query)
	if err != nil {
		slog.Error("failed to generate query embedding", "error", err)
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}
	if len(embedding) != expectedEmbeddingDimensions {
		return nil, fmt.Errorf("unexpected embedding size: got %d want %d", len(embedding), expectedEmbeddingDimensions)
	}

	filterSQL := GenerateFilterSQL(input)

	slog.Info("search query built",
		"query", input.Query,
		"embedding_dim", len(embedding),
		"tenant_id", input.TenantID,
		"filter_sql", filterSQL.WhereClause,
	)

	chunkLimit := expandChunkCandidateLimit(limit)

	repoReq := repository.SearchRequest{
		Query:             input.Query,
		TenantID:          input.TenantID,
		OrgID:             input.OrgID,
		Limit:             chunkLimit,
		FilterMeta:        map[string]interface{}{},
		QueryVector:       search.FormatVectorLiteral(embedding),
		FilterWhereClause: filterSQL.WhereClause,
		FilterParams:      filterSQL.Params,
		MinScore:          minimumSearchScore,
	}

	repoResult, err := s.searchRepo.Search(ctx, repoReq)
	if err != nil {
		slog.Error("search repository failed", "error", err)
		return nil, fmt.Errorf("search repository failed: %w", err)
	}

	aggregated := make(map[string]aggregatedSearchResult, len(repoResult.Chunks))
	order := make([]string, 0, len(repoResult.Chunks))
	for rank, chunk := range repoResult.Chunks {
		result, exists := aggregated[chunk.DocumentID]
		if !exists {
			file := chunk.DocumentTitle
			if file == "" {
				file = chunk.DocumentID
			}

			aggregated[chunk.DocumentID] = aggregatedSearchResult{
				SearchResult: SearchResult{
					DocumentID: chunk.DocumentID,
					File:       file,
					MaxScore:   chunk.Score,
				},
				rankScore: reciprocalRank(rank),
			}
			order = append(order, chunk.DocumentID)
			continue
		}

		result.rankScore += reciprocalRank(rank)
		if chunk.Score > result.MaxScore {
			result.MaxScore = chunk.Score
		}
		if result.File == "" && chunk.DocumentTitle != "" {
			result.File = chunk.DocumentTitle
		}

		aggregated[chunk.DocumentID] = result
	}

	results := make([]aggregatedSearchResult, 0, len(aggregated))
	for _, documentID := range order {
		result := aggregated[documentID]
		if result.File == "" {
			result.File = result.DocumentID
		}
		results = append(results, result)
	}

	sort.SliceStable(results, func(i, j int) bool {
		if results[i].MaxScore == results[j].MaxScore {
			if results[i].rankScore == results[j].rankScore {
				return results[i].DocumentID < results[j].DocumentID
			}
			return results[i].rankScore > results[j].rankScore
		}
		return results[i].MaxScore > results[j].MaxScore
	})

	if len(results) > limit {
		results = results[:limit]
	}

	fileResults := make([]SearchResult, 0, len(results))
	for _, result := range results {
		fileResults = append(fileResults, result.SearchResult)
	}

	page := 1
	pageSize := limit
	totalPages := 1
	if len(fileResults) > 0 {
		totalPages = (len(fileResults) + pageSize - 1) / pageSize
	}
	hasMore := page < totalPages

	var nextCursor string
	if hasMore && len(fileResults) > 0 {
		last := fileResults[len(fileResults)-1]
		nextCursor = last.DocumentID
	}

	return &SearchOutput{
		Results:    fileResults,
		Query:      input.Query,
		Total:      len(fileResults),
		Page:       page,
		PageSize:   pageSize,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}, nil
}

func (s *SearchService) SearchWithFilters(ctx context.Context, input *SearchInput) (*SearchOutput, error) {
	return s.Search(ctx, input)
}
