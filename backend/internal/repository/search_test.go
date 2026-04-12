package repository

import (
	"testing"
)

func TestBuildSearchQuery_UsesVectorSimilarityAndThreshold(t *testing.T) {
	query, args, err := buildSearchQuery(SearchRequest{
		Query:             "PDFObject",
		QueryVector:       "[0.25,0.75]",
		FilterWhereClause: "d.tenant_id = $2 AND d.org_id = $3",
		FilterParams:      []interface{}{"tenant-1", "org-1"},
		MinScore:          0.01,
		Limit:             10,
	})
	if err != nil {
		t.Fatalf("buildSearchQuery() error = %v", err)
	}

	assertContains(t, query, "WITH vector_matches AS")
	assertContains(t, query, "WITH vector_matches AS")
	assertContains(t, query, "POSITION(LOWER($4) IN LOWER(c.chunk_text)) > 0")
	assertContains(t, query, "LEAST(1.0, GREATEST(0.0")
	assertContains(t, query, "WHERE c.embedding IS NOT NULL")
	assertContains(t, query, "d.tenant_id = $2 AND d.org_id = $3")
	assertContains(t, query, "WHERE score > $5")
	assertContains(t, query, "ORDER BY score DESC LIMIT 10")

	if len(args) != 5 {
		t.Fatalf("len(args) = %d, want 5", len(args))
	}
	if args[0] != "[0.25,0.75]" {
		t.Fatalf("args[0] = %v, want query vector", args[0])
	}
	if args[3] != "PDFObject" {
		t.Fatalf("args[3] = %v, want PDFObject", args[3])
	}
	if args[4] != 0.01 {
		t.Fatalf("args[4] = %v, want 0.01", args[4])
	}
}

func TestBuildSearchQuery_RequiresVector(t *testing.T) {
	_, _, err := buildSearchQuery(SearchRequest{Query: "test"})
	if err == nil {
		t.Fatal("expected error when query vector is missing")
	}
}

func TestBuildSearchQuery_RequiresQueryText(t *testing.T) {
	_, _, err := buildSearchQuery(SearchRequest{QueryVector: "[0.25,0.75]"})
	if err == nil {
		t.Fatal("expected error when query text is missing")
	}
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !contains(s, substr) {
		t.Fatalf("expected %q to contain %q", s, substr)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
