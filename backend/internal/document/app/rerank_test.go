package app

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/docvault/backend/internal/repository"
)

func chunk(id, text string, score float64) repository.ChunkMatch {
	return repository.ChunkMatch{ChunkID: id, ChunkText: text, Score: score}
}

func idsOf(chunks []repository.ChunkMatch) []string {
	out := make([]string, len(chunks))
	for i, c := range chunks {
		out[i] = c.ChunkID
	}
	return out
}

func TestRerankAndTrim_ShortSetIsUntouched(t *testing.T) {
	in := []repository.ChunkMatch{chunk("a", "alpha", 0.9), chunk("b", "beta", 0.8)}
	// A reranker that would reorder — it must NOT be called for a set <= limit.
	rk := &scriptedReranker{scores: []RerankScore{{Index: 1, Score: 1}, {Index: 0, Score: 0}}}

	got := rerankAndTrim(context.Background(), rk, "q", in, 5)
	if !reflect.DeepEqual(idsOf(got), []string{"a", "b"}) {
		t.Fatalf("short set reordered unexpectedly: %v", idsOf(got))
	}
	if rk.calls != 0 {
		t.Fatalf("reranker called %d time(s) on a short set; want 0", rk.calls)
	}
}

func TestRerankAndTrim_ReordersAndTrims(t *testing.T) {
	in := []repository.ChunkMatch{
		chunk("a", "alpha", 0.9),
		chunk("b", "beta", 0.8),
		chunk("c", "gamma", 0.7),
	}
	// Promote index 2 ("gamma") to the top, drop index 1 ("beta") via the limit.
	rk := &scriptedReranker{scores: []RerankScore{
		{Index: 2, Score: 0.99},
		{Index: 0, Score: 0.5},
		{Index: 1, Score: 0.1},
	}}

	got := rerankAndTrim(context.Background(), rk, "q", in, 2)
	if !reflect.DeepEqual(idsOf(got), []string{"c", "a"}) {
		t.Fatalf("rerank+trim order = %v, want [c a]", idsOf(got))
	}
}

func TestRerankAndTrim_ErrorFallsBackToSQLOrder(t *testing.T) {
	in := []repository.ChunkMatch{
		chunk("a", "alpha", 0.9),
		chunk("b", "beta", 0.8),
		chunk("c", "gamma", 0.7),
	}
	rk := &scriptedReranker{err: errors.New("sidecar down")}

	got := rerankAndTrim(context.Background(), rk, "q", in, 2)
	// Falls back to SQL order (already score-desc) trimmed to limit.
	if !reflect.DeepEqual(idsOf(got), []string{"a", "b"}) {
		t.Fatalf("error fallback = %v, want [a b]", idsOf(got))
	}
}

func TestRerankAndTrim_NoopPreservesOrder(t *testing.T) {
	in := []repository.ChunkMatch{
		chunk("a", "alpha", 0.9),
		chunk("b", "beta", 0.8),
		chunk("c", "gamma", 0.7),
	}
	got := rerankAndTrim(context.Background(), NoopReranker{}, "q", in, 2)
	if !reflect.DeepEqual(idsOf(got), []string{"a", "b"}) {
		t.Fatalf("noop order = %v, want [a b]", idsOf(got))
	}
}

func TestHTTPReranker_ParsesTEIResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rerank" || r.Method != http.MethodPost {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		var req struct {
			Query string   `json:"query"`
			Texts []string `json:"texts"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Query != "teepee" || len(req.Texts) != 2 {
			t.Errorf("request = %+v", req)
		}
		// TEI returns results sorted by score desc, each with the original index.
		_ = json.NewEncoder(w).Encode([]teiRerankItem{
			{Index: 1, Score: 0.98},
			{Index: 0, Score: 0.21},
		})
	}))
	defer srv.Close()

	rk := NewHTTPReranker(srv.URL, srv.Client())
	scores, err := rk.Rerank(context.Background(), "teepee", []string{"alpha", "beta"})
	if err != nil {
		t.Fatalf("Rerank error = %v", err)
	}
	if len(scores) != 2 || scores[0].Index != 1 || scores[1].Index != 0 {
		t.Fatalf("scores = %+v", scores)
	}
}

func TestHTTPReranker_EmptyBaseURLIsNoop(t *testing.T) {
	rk := NewHTTPReranker("", nil)
	scores, err := rk.Rerank(context.Background(), "q", []string{"a"})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if scores != nil {
		t.Fatalf("expected nil scores for empty base URL, got %+v", scores)
	}
}

// scriptedReranker returns canned scores (or an error) and counts calls.
type scriptedReranker struct {
	scores []RerankScore
	err    error
	calls  int
}

func (s *scriptedReranker) Rerank(_ context.Context, _ string, _ []string) ([]RerankScore, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return s.scores, nil
}

var _ RerankerPort = (*scriptedReranker)(nil)
var _ RerankerPort = NoopReranker{}
