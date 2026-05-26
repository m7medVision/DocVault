// Package usecase provides tests for DocVault use cases.
package usecase

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

func TestSearchInput_Validation(t *testing.T) {
	svc := NewSearchService(nil, nil)
	_, err := svc.Search(context.Background(), &SearchInput{Query: "", TenantID: "tenant-1"})
	if err == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestSearch_PassesTypedFiltersToRepository(t *testing.T) {
	embedding := make([]float32, 1024)
	embedding[0] = 0.25
	embedding[1] = 0.75

	repo := &stubSearchRepository{result: &repository.SearchResult{}}
	svc := NewSearchService(stubEmbedder{embedding: embedding}, repo)

	_, err := svc.Search(context.Background(), &SearchInput{
		Query:     "PDFObject",
		TenantID:  "tenant-1",
		OrgID:     "org-1",
		DocType:   "contract",
		Language:  "en",
		Status:    "processed",
		FolderID:  "folder-1",
		Tags:      []string{"legal"},
		StartDate: "2024-01-01",
		EndDate:   "2024-12-31",
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if repo.lastReq.QueryVector == "" {
		t.Fatal("expected query vector to be passed to repository")
	}
	if repo.lastReq.Limit != 50 {
		t.Fatalf("limit = %d, want 50", repo.lastReq.Limit)
	}
	if repo.lastReq.TenantID != "tenant-1" || repo.lastReq.OrgID != "org-1" {
		t.Fatalf("tenant/org = %s/%s", repo.lastReq.TenantID, repo.lastReq.OrgID)
	}
	if repo.lastReq.DocType != "contract" || repo.lastReq.Language != "en" || repo.lastReq.Status != "processed" || repo.lastReq.FolderID != "folder-1" {
		t.Fatalf("filters not passed: %#v", repo.lastReq)
	}
	if len(repo.lastReq.Tags) != 1 || repo.lastReq.Tags[0] != "legal" {
		t.Fatalf("tags = %#v", repo.lastReq.Tags)
	}
	if repo.lastReq.StartDate == nil || repo.lastReq.EndDate == nil {
		t.Fatalf("expected parsed date filters: %#v", repo.lastReq)
	}
	if repo.lastReq.MinScore != minimumSearchScore {
		t.Fatalf("min_score = %v, want %v", repo.lastReq.MinScore, minimumSearchScore)
	}
}

func TestSearch_RejectsInvalidDateFilter(t *testing.T) {
	embedding := make([]float32, 1024)
	repo := &stubSearchRepository{result: &repository.SearchResult{}}
	svc := NewSearchService(stubEmbedder{embedding: embedding}, repo)

	_, err := svc.Search(context.Background(), &SearchInput{
		Query:     "PDFObject",
		TenantID:  "tenant-1",
		OrgID:     "org-1",
		StartDate: "not-a-date",
	})
	if err == nil {
		t.Fatal("expected invalid date error")
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
