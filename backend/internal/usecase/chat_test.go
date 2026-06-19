package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/docvault/backend/internal/repository"
)

// fixedEmbedding returns a 1024-length vector so it passes the dimension check.
func fixedEmbedding() []float32 {
	embedding := make([]float32, expectedEmbeddingDimensions)
	embedding[0] = 0.5
	embedding[1] = 0.25
	return embedding
}

// stubChunks is a fixed set of chunks with distinct document/title/page/score.
func stubChunks() []repository.ChunkMatch {
	return []repository.ChunkMatch{
		{
			DocumentID:    "doc-alpha",
			DocumentTitle: "Alpha Report",
			ChunkID:       "chunk-1",
			ChunkText:     "The alpha figure was 42 units.",
			PageNumber:    3,
			Score:         0.91,
		},
		{
			DocumentID:    "doc-beta",
			DocumentTitle: "Beta Memo",
			ChunkID:       "chunk-2",
			ChunkText:     "The beta target is 7 percent.",
			PageNumber:    11,
			Score:         0.77,
		},
	}
}

func TestBuildRetrievalContext_NumbersContextAndMapsSources(t *testing.T) {
	chunks := stubChunks()

	contextBlock, sources := buildRetrievalContext(chunks)

	// Numbered context contains [1].. markers, titles, pages, and chunk text.
	for _, want := range []string{
		"[1]", "[2]",
		"Alpha Report", "Beta Memo",
		"page 3", "page 11",
		"The alpha figure was 42 units.",
		"The beta target is 7 percent.",
	} {
		if !strings.Contains(contextBlock, want) {
			t.Fatalf("context block missing %q\n---\n%s", want, contextBlock)
		}
	}

	// Sources map 1:1 with the chunks.
	if len(sources) != len(chunks) {
		t.Fatalf("len(sources) = %d, want %d", len(sources), len(chunks))
	}
	for i, src := range sources {
		chunk := chunks[i]
		if src.N != i+1 {
			t.Fatalf("sources[%d].N = %d, want %d", i, src.N, i+1)
		}
		if src.DocumentID != chunk.DocumentID {
			t.Fatalf("sources[%d].DocumentID = %q, want %q", i, src.DocumentID, chunk.DocumentID)
		}
		if src.Title != chunk.DocumentTitle {
			t.Fatalf("sources[%d].Title = %q, want %q", i, src.Title, chunk.DocumentTitle)
		}
		if src.Page != chunk.PageNumber {
			t.Fatalf("sources[%d].Page = %d, want %d", i, src.Page, chunk.PageNumber)
		}
		if src.Score != chunk.Score {
			t.Fatalf("sources[%d].Score = %v, want %v", i, src.Score, chunk.Score)
		}
	}
}

func TestBuildRetrievalContext_EmptyChunks(t *testing.T) {
	contextBlock, sources := buildRetrievalContext(nil)
	if contextBlock != "" {
		t.Fatalf("context block = %q, want empty", contextBlock)
	}
	if sources != nil {
		t.Fatalf("sources = %#v, want nil", sources)
	}
}

// cannedChatServer returns an httptest.Server that responds with a minimal
// OpenRouter-style SSE stream.
func cannedChatServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/chat/completions") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		lines := []string{
			`data: {"choices":[{"delta":{"content":"The alpha figure "}}]}`,
			`data: {"choices":[{"delta":{"content":"was 42 units [1]."}}]}`,
			`data: [DONE]`,
		}
		for _, line := range lines {
			fmt.Fprintf(w, "%s\n\n", line)
		}
	}))
}

func TestStreamChat_EmitsContentAndSourcesAndFinishes(t *testing.T) {
	server := cannedChatServer(t)
	defer server.Close()

	repo := &stubSearchRepository{result: &repository.SearchResult{Chunks: stubChunks()}}
	svc := NewChatService(stubEmbedder{embedding: fixedEmbedding()}, repo)
	svc.chatBaseURL = server.URL
	svc.httpClient = server.Client()

	var buf bytes.Buffer
	err := svc.StreamChat(context.Background(), &ChatInput{
		Messages:  []ChatMessage{{Role: "user", Content: "what was the alpha figure?"}},
		TenantID:  "tenant-1",
		OrgID:     "org-1",
		APIKey:    "test-key",
		ChatModel: "test-model",
	}, &buf)
	if err != nil {
		t.Fatalf("StreamChat() error = %v", err)
	}

	// Retrieval must be scoped to the caller's tenant/org. Global chat (no
	// DocumentID) must not constrain to a single document. This is the guarantee
	// that keeps chat from leaking across tenants/orgs.
	if repo.lastReq.TenantID != "tenant-1" {
		t.Fatalf("retrieval TenantID = %q, want %q", repo.lastReq.TenantID, "tenant-1")
	}
	if repo.lastReq.OrgID != "org-1" {
		t.Fatalf("retrieval OrgID = %q, want %q", repo.lastReq.OrgID, "org-1")
	}
	if repo.lastReq.DocumentID != "" {
		t.Fatalf("global chat DocumentID = %q, want empty", repo.lastReq.DocumentID)
	}
	if repo.lastReq.MinScore != minimumSearchScore {
		t.Fatalf("retrieval MinScore = %v, want %v", repo.lastReq.MinScore, minimumSearchScore)
	}

	out := buf.String()

	if !strings.Contains(out, "TEXT_MESSAGE_CONTENT") {
		t.Fatalf("output missing TEXT_MESSAGE_CONTENT\n---\n%s", out)
	}

	sources := extractSourcesEvent(t, out)
	if sources == nil {
		t.Fatalf("output missing SOURCES event\n---\n%s", out)
	}
	if len(sources) != 2 {
		t.Fatalf("SOURCES len = %d, want 2", len(sources))
	}
	assertSourceMatch(t, sources, "doc-alpha", 3)
	assertSourceMatch(t, sources, "doc-beta", 11)

	// RUN_FINISHED must be the terminal event.
	if !strings.Contains(out, "RUN_FINISHED") {
		t.Fatalf("output missing RUN_FINISHED\n---\n%s", out)
	}
	finishIdx := strings.LastIndex(out, "RUN_FINISHED")
	sourcesIdx := strings.LastIndex(out, `"SOURCES"`)
	if finishIdx < sourcesIdx {
		t.Fatalf("RUN_FINISHED must come after SOURCES\n---\n%s", out)
	}
}

func TestStreamChat_EmptyRetrievalNoSources(t *testing.T) {
	// Server should never be called; if it is, fail loudly.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("chat endpoint should not be called when there is no grounding")
	}))
	defer server.Close()

	repo := &stubSearchRepository{result: &repository.SearchResult{Chunks: nil}}
	svc := NewChatService(stubEmbedder{embedding: fixedEmbedding()}, repo)
	svc.chatBaseURL = server.URL
	svc.httpClient = server.Client()

	var buf bytes.Buffer
	err := svc.StreamChat(context.Background(), &ChatInput{
		Messages:  []ChatMessage{{Role: "user", Content: "anything in here?"}},
		TenantID:  "tenant-1",
		OrgID:     "org-1",
		APIKey:    "test-key",
		ChatModel: "test-model",
	}, &buf)
	if err != nil {
		t.Fatalf("StreamChat() error = %v", err)
	}

	out := buf.String()

	if strings.Contains(out, `"SOURCES"`) {
		t.Fatalf("expected NO SOURCES event for empty retrieval\n---\n%s", out)
	}
	if !strings.Contains(out, "TEXT_MESSAGE_CONTENT") {
		t.Fatalf("expected a graceful message to be streamed\n---\n%s", out)
	}
	if !strings.Contains(out, "couldn't find") {
		t.Fatalf("expected a graceful no-results message\n---\n%s", out)
	}
	if !strings.Contains(out, "RUN_FINISHED") {
		t.Fatalf("output missing RUN_FINISHED\n---\n%s", out)
	}
}

func TestStreamChat_PropagatesDocumentScope(t *testing.T) {
	server := cannedChatServer(t)
	defer server.Close()

	repo := &stubSearchRepository{result: &repository.SearchResult{Chunks: stubChunks()}}
	svc := NewChatService(stubEmbedder{embedding: fixedEmbedding()}, repo)
	svc.chatBaseURL = server.URL
	svc.httpClient = server.Client()

	var buf bytes.Buffer
	err := svc.StreamChat(context.Background(), &ChatInput{
		DocumentID: "doc-scoped",
		Messages:   []ChatMessage{{Role: "user", Content: "what does this say?"}},
		TenantID:   "tenant-1",
		OrgID:      "org-1",
		APIKey:     "test-key",
		ChatModel:  "test-model",
	}, &buf)
	if err != nil {
		t.Fatalf("StreamChat() error = %v", err)
	}

	if repo.lastReq.DocumentID != "doc-scoped" {
		t.Fatalf("per-document chat DocumentID = %q, want %q", repo.lastReq.DocumentID, "doc-scoped")
	}
	if repo.lastReq.TenantID != "tenant-1" || repo.lastReq.OrgID != "org-1" {
		t.Fatalf("per-document chat must stay tenant/org scoped; got tenant=%q org=%q", repo.lastReq.TenantID, repo.lastReq.OrgID)
	}
}

// extractSourcesEvent parses each SSE "data: " line and returns the sources
// array from the first SOURCES event, or nil if none is present.
func extractSourcesEvent(t *testing.T, out string) []ChatSource {
	t.Helper()
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		data, ok := strings.CutPrefix(line, "data: ")
		if !ok {
			continue
		}
		var event struct {
			Type    string       `json:"type"`
			Sources []ChatSource `json:"sources"`
		}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}
		if event.Type == "SOURCES" {
			return event.Sources
		}
	}
	return nil
}

func assertSourceMatch(t *testing.T, sources []ChatSource, documentID string, page int) {
	t.Helper()
	for _, src := range sources {
		if src.DocumentID == documentID && src.Page == page {
			return
		}
	}
	t.Fatalf("SOURCES missing documentId=%q page=%d; got %#v", documentID, page, sources)
}
