package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"

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
	UserID    string
	GroupIDs  []string
	IsAdmin   bool
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

	startDate, err := parseSearchDate(input.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date: %w", err)
	}
	endDate, err := parseSearchDate(input.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date: %w", err)
	}

	slog.Info("search query built",
		"query", input.Query,
		"embedding_dim", len(embedding),
		"tenant_id", input.TenantID,
		"org_id", input.OrgID,
	)

	chunkLimit := expandChunkCandidateLimit(limit)

	repoReq := repository.SearchRequest{
		Query:       input.Query,
		TenantID:    input.TenantID,
		OrgID:       input.OrgID,
		UserID:      input.UserID,
		GroupIDs:    input.GroupIDs,
		IsAdmin:     input.IsAdmin,
		DocType:     input.DocType,
		Language:    input.Language,
		Status:      input.Status,
		FolderID:    input.FolderID,
		Tags:        input.Tags,
		StartDate:   startDate,
		EndDate:     endDate,
		Limit:       chunkLimit,
		QueryVector: search.FormatVectorLiteral(embedding),
		MinScore:    minimumSearchScore,
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

func parseSearchDate(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return &parsed, nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func (s *SearchService) SearchWithFilters(ctx context.Context, input *SearchInput) (*SearchOutput, error) {
	return s.Search(ctx, input)
}
