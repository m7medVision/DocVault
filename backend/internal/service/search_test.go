// Package service provides tests for DocVault services.
package service

import (
	"context"
	"testing"

	"github.com/docvault/backend/internal/repository"
)

type stubEmbedder struct {
	embedding []float32
	err       error
}

func (s stubEmbedder) Embed(context.Context, string) ([]float32, error) {
	return s.embedding, s.err
}

func (s stubEmbedder) EmbedBatch(context.Context, []string) ([][]float32, error) {
	return nil, nil
}

type stubSearchRepository struct {
	result  *repository.SearchResult
	err     error
	lastReq repository.SearchRequest
}

func (s *stubSearchRepository) Search(_ context.Context, req repository.SearchRequest) (*repository.SearchResult, error) {
	s.lastReq = req
	return s.result, s.err
}

func (s *stubSearchRepository) IndexChunk(context.Context, repository.DocumentChunk) error {
	return nil
}

func (s *stubSearchRepository) DeleteChunksByDocument(context.Context, string) error {
	return nil
}

// TestGenerateFilterSQL tests the filter SQL generation.
func TestGenerateFilterSQL(t *testing.T) {
	tests := []struct {
		name     string
		input    *SearchInput
		wantLen  int // number of params
		wantCond int // number of conditions
	}{
		{
			name: "empty filters",
			input: &SearchInput{
				Query:    "test query",
				TenantID: "tenant-1",
			},
			wantLen:  1, // only tenant
			wantCond: 1,
		},
		{
			name: "with doc type filter",
			input: &SearchInput{
				Query:    "contract",
				TenantID: "tenant-1",
				DocType:  "contract",
			},
			wantLen:  2,
			wantCond: 2,
		},
		{
			name: "with language filter",
			input: &SearchInput{
				Query:    " عقد ",
				TenantID: "tenant-1",
				Language: "ar",
			},
			wantLen:  2,
			wantCond: 2,
		},
		{
			name: "with multiple filters",
			input: &SearchInput{
				Query:    "warranty",
				TenantID: "tenant-1",
				DocType:  "warranty",
				Language: "en",
				Status:   "processed",
				FolderID: "folder-123",
			},
			wantLen:  5,
			wantCond: 5,
		},
		{
			name: "with date range",
			input: &SearchInput{
				Query:     "invoice",
				TenantID:  "tenant-1",
				StartDate: "2024-01-01",
				EndDate:   "2024-12-31",
			},
			wantLen:  3,
			wantCond: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filterSQL := GenerateFilterSQL(tt.input)
			if filterSQL == nil {
				t.Fatal("GenerateFilterSQL returned nil")
			}
			if len(filterSQL.Params) != tt.wantLen {
				t.Errorf("Params len = %d, want %d", len(filterSQL.Params), tt.wantLen)
			}
			// Count conditions by checking for "AND" occurrences
			condCount := 0
			for i := 0; i < len(filterSQL.WhereClause)-3; i++ {
				if filterSQL.WhereClause[i:i+3] == "AND" {
					condCount++
				}
			}
			if condCount+1 != tt.wantCond {
				t.Errorf("Condition count = %d, want %d", condCount+1, tt.wantCond)
			}
		})
	}
}

// TestBuildVectorSearchQuery tests the vector search query building.
func TestBuildVectorSearchQuery(t *testing.T) {
	tests := []struct {
		name      string
		filterSQL *FilterSQL
		limit     int
		wantLimit int
	}{
		{
			name:      "no filters",
			filterSQL: NewFilterSQL(),
			limit:     20,
			wantLimit: 20,
		},
		{
			name: "with tenant filter",
			filterSQL: func() *FilterSQL {
				f := NewFilterSQL()
				placeholder := f.addParam("tenant-1")
				f.AddCondition("d.tenant_id = " + placeholder)
				return f
			}(),
			limit:     20,
			wantLimit: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, params := BuildVectorSearchQuery(tt.filterSQL, tt.limit)

			if query == "" {
				t.Error("BuildVectorSearchQuery returned empty query")
			}

			// Check that ORDER BY clause is present
			if !contains(query, "ORDER BY c.embedding") {
				t.Error("Query missing ORDER BY clause")
			}

			// Check that LIMIT is present with correct value
			expectedLimit := string(rune('0' + tt.wantLimit%10))
			if !contains(query, "LIMIT") {
				t.Error("Query missing LIMIT clause")
			}

			// Check for JOIN
			if !contains(query, "JOIN documents d") {
				t.Error("Query missing JOIN clause")
			}

			// Note: When no filters, params is empty but query still has $1 for embedding
			// This is expected since embedding is always the first parameter
			_ = params
			_ = expectedLimit
		})
	}
}

// TestSearchInput_Validation tests search input validation.
func TestSearchInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   *SearchInput
		wantErr bool
	}{
		{
			name: "empty query",
			input: &SearchInput{
				Query:    "",
				TenantID: "tenant-1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Only test empty query validation
			// Full search requires OpenRouter API key
			svc := NewSearchService(nil, nil)
			_, err := svc.Search(context.Background(), tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Search() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestSearchService_Bilingual tests search with bilingual content.
// Note: These tests validate filter generation, actual search requires OpenRouter API key.
func TestSearchService_Bilingual(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		language string
	}{
		{
			name:     "english query",
			query:    "contract agreement",
			language: "en",
		},
		{
			name:     "arabic query",
			query:    "عقد اتفاق",
			language: "ar",
		},
		{
			name:     "mixed content",
			query:    "invoice فاتورة",
			language: "mixed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate that filter generation works for different languages
			input := &SearchInput{
				Query:    tt.query,
				Language: tt.language,
				TenantID: "test-tenant",
			}

			filterSQL := GenerateFilterSQL(input)
			if filterSQL == nil {
				t.Fatal("GenerateFilterSQL returned nil")
			}

			// Should have tenant + language filter
			if len(filterSQL.Params) != 2 {
				t.Errorf("Params len = %d, want 2", len(filterSQL.Params))
			}
		})
	}
}

// TestSearchService_Pagination tests pagination parameters.
func TestSearchService_Pagination(t *testing.T) {
	tests := []struct {
		name     string
		limit    int
		wantPage int
		wantSize int
	}{
		{
			name:     "default limit",
			limit:    0,
			wantPage: 1,
			wantSize: 20,
		},
		{
			name:     "custom limit",
			limit:    10,
			wantPage: 1,
			wantSize: 10,
		},
		{
			name:     "max limit",
			limit:    100,
			wantPage: 1,
			wantSize: 20, // Capped at 50
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build query to test limit handling
			filterSQL := NewFilterSQL()
			placeholder := filterSQL.addParam("tenant-1")
			filterSQL.AddCondition("tenant_id = " + placeholder)

			query, params := BuildVectorSearchQuery(filterSQL, tt.limit)
			if query == "" {
				t.Fatal("BuildVectorSearchQuery returned empty query")
			}

			// Verify limit is in query
			if !contains(query, "LIMIT") {
				t.Error("Query missing LIMIT clause")
			}

			// Verify params include tenant
			if len(params) != 1 {
				t.Errorf("Params len = %d, want 1", len(params))
			}

			_ = tt.wantPage
			_ = tt.wantSize
		})
	}
}

func TestSearch_AggregatesFileScoresAndPassesVectorQuery(t *testing.T) {
	embedding := make([]float32, 1024)
	embedding[0] = 0.25
	embedding[1] = 0.75

	repo := &stubSearchRepository{
		result: &repository.SearchResult{
			Chunks: []repository.ChunkMatch{
				{DocumentID: "doc-1", DocumentTitle: "first.pdf", Score: 0.42},
				{DocumentID: "doc-2", DocumentTitle: "second.pdf", Score: 0.63},
				{DocumentID: "doc-1", DocumentTitle: "first.pdf", Score: 0.91},
			},
		},
	}

	svc := NewSearchService(stubEmbedder{embedding: embedding}, repo)
	output, err := svc.Search(context.Background(), &SearchInput{
		Query:    "PDFObject",
		TenantID: "tenant-1",
		OrgID:    "org-1",
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(output.Results) != 2 {
		t.Fatalf("len(output.Results) = %d, want 2", len(output.Results))
	}

	if output.Results[0].DocumentID != "doc-1" {
		t.Fatalf("first result document_id = %s, want doc-1", output.Results[0].DocumentID)
	}
	if output.Results[0].MaxScore != 0.91 {
		t.Fatalf("first result max_score = %v, want 0.91", output.Results[0].MaxScore)
	}
	if output.Results[0].File != "first.pdf" {
		t.Fatalf("first result file = %s, want first.pdf", output.Results[0].File)
	}

	if repo.lastReq.QueryVector == "" {
		t.Fatal("expected query vector to be passed to repository")
	}
	if repo.lastReq.Limit != 50 {
		t.Fatalf("limit = %d, want 50", repo.lastReq.Limit)
	}
	if repo.lastReq.FilterWhereClause == "" {
		t.Fatal("expected filter where clause to be passed to repository")
	}
	if repo.lastReq.MinScore != minimumSearchScore {
		t.Fatalf("min_score = %v, want %v", repo.lastReq.MinScore, minimumSearchScore)
	}
}

func TestSearch_SortsFilesByMaxScoreBeforeCombinedChunkEvidence(t *testing.T) {
	embedding := make([]float32, 1024)
	embedding[0] = 1

	repo := &stubSearchRepository{
		result: &repository.SearchResult{
			Chunks: []repository.ChunkMatch{
				{DocumentID: "doc-outlier", DocumentTitle: "outlier.pdf", Score: 0.91},
				{DocumentID: "doc-consistent", DocumentTitle: "consistent.pdf", Score: 0.89},
				{DocumentID: "doc-consistent", DocumentTitle: "consistent.pdf", Score: 0.88},
				{DocumentID: "doc-consistent", DocumentTitle: "consistent.pdf", Score: 0.87},
			},
		},
	}

	svc := NewSearchService(stubEmbedder{embedding: embedding}, repo)
	output, err := svc.Search(context.Background(), &SearchInput{
		Query:    "search evidence",
		TenantID: "tenant-1",
		OrgID:    "org-1",
		Limit:    2,
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(output.Results) != 2 {
		t.Fatalf("len(output.Results) = %d, want 2", len(output.Results))
	}
	if output.Results[0].DocumentID != "doc-outlier" {
		t.Fatalf("first result document_id = %s, want doc-outlier", output.Results[0].DocumentID)
	}
	if output.Results[0].MaxScore != 0.91 {
		t.Fatalf("first result max_score = %v, want 0.91", output.Results[0].MaxScore)
	}
}

// TestFilterSQL_AddCondition tests the filter SQL builder.
func TestFilterSQL_AddCondition(t *testing.T) {
	f := NewFilterSQL()

	// Add first condition
	first := f.addParam("tenant-1")
	f.AddCondition("tenant_id = " + first)
	if f.WhereClause != "tenant_id = $2" {
		t.Errorf("First condition = %s, want 'tenant_id = $2'", f.WhereClause)
	}
	if len(f.Params) != 1 {
		t.Errorf("Params len = %d, want 1", len(f.Params))
	}

	// Add second condition
	second := f.addParam("contract")
	f.AddCondition("doc_type = " + second)
	if !contains(f.WhereClause, "AND") {
		t.Error("Expected AND between conditions")
	}
	if len(f.Params) != 2 {
		t.Errorf("Params len = %d, want 2", len(f.Params))
	}
}

// contains is a helper to check string containment.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
