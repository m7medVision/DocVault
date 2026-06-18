package repository

import (
	"testing"
	"time"
)

func TestSearchDocumentChunksParams_MapsTypedFilters(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	params, err := searchDocumentChunksParams(SearchRequest{
		Query:       "PDFObject",
		QueryVector: "[0.25,0.75]",
		TenantID:    "tenant-1",
		UserID:      "user-1",
		OrgID:       "org-1",
		DocType:     "contract",
		Language:    "en",
		Status:      "processed",
		FolderID:    "folder-1",
		Tags:        []string{"legal", "signed"},
		StartDate:   &start,
		EndDate:     &end,
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("searchDocumentChunksParams() error = %v", err)
	}

	if params.QueryVector != "[0.25,0.75]" {
		t.Fatalf("query vector = %q", params.QueryVector)
	}
	if params.QueryText != "PDFObject" {
		t.Fatalf("query text = %q", params.QueryText)
	}
	if params.LimitCount != 10 {
		t.Fatalf("limit = %d, want 10", params.LimitCount)
	}
	if params.DocType == nil || *params.DocType != "contract" {
		t.Fatalf("doc type = %v", params.DocType)
	}
	if params.Language == nil || *params.Language != "en" {
		t.Fatalf("language = %v", params.Language)
	}
	if params.Status == nil || *params.Status != "processed" {
		t.Fatalf("status = %v", params.Status)
	}
	if params.FolderID == nil || *params.FolderID != "folder-1" {
		t.Fatalf("folder = %v", params.FolderID)
	}
	if len(params.Tags) != 2 || params.Tags[0] != "legal" || params.Tags[1] != "signed" {
		t.Fatalf("tags = %#v", params.Tags)
	}
	if !params.StartDate.Valid || !params.StartDate.Time.Equal(start) {
		t.Fatalf("start date = %#v", params.StartDate)
	}
	if !params.EndDate.Valid || !params.EndDate.Time.Equal(end) {
		t.Fatalf("end date = %#v", params.EndDate)
	}
}

func TestSearchDocumentChunksParams_DefaultsAndValidation(t *testing.T) {
	t.Run("requires vector", func(t *testing.T) {
		_, err := searchDocumentChunksParams(SearchRequest{Query: "test"})
		if err == nil {
			t.Fatal("expected error when query vector is missing")
		}
	})

	t.Run("requires query text", func(t *testing.T) {
		_, err := searchDocumentChunksParams(SearchRequest{QueryVector: "[0.25,0.75]"})
		if err == nil {
			t.Fatal("expected error when query text is missing")
		}
	})

	t.Run("requires user id", func(t *testing.T) {
		_, err := searchDocumentChunksParams(SearchRequest{Query: "test", QueryVector: "[0.25,0.75]"})
		if err == nil {
			t.Fatal("expected error when user id is missing")
		}
	})

	t.Run("defaults invalid limit and empty tags", func(t *testing.T) {
		params, err := searchDocumentChunksParams(SearchRequest{Query: "test", QueryVector: "[0.25,0.75]", UserID: "user-1", Limit: 100})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if params.LimitCount != 20 {
			t.Fatalf("limit = %d, want 20", params.LimitCount)
		}
		if params.Tags == nil || len(params.Tags) != 0 {
			t.Fatalf("tags = %#v, want empty slice", params.Tags)
		}
		if params.GroupIds == nil || len(params.GroupIds) != 0 {
			t.Fatalf("group ids = %#v, want empty slice", params.GroupIds)
		}
	})
}
