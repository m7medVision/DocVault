package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/docvault/backend/internal/repository"
)

// RerankScore maps a document (by its original index in the input slice) to a
// cross-encoder relevance score.
type RerankScore struct {
	Index int
	Score float64
}

// RerankerPort scores documents against a query with a cross-encoder reranker.
// Rerank returns one RerankScore per input document.
type RerankerPort interface {
	Rerank(ctx context.Context, query string, documents []string) ([]RerankScore, error)
}

// NoopReranker is the default when no rerank endpoint is configured. It returns
// stable identity scores so a sort-by-score preserves the SQL retrieval order
// and chat keeps working without a sidecar.
type NoopReranker struct{}

func (NoopReranker) Rerank(_ context.Context, _ string, documents []string) ([]RerankScore, error) {
	out := make([]RerankScore, len(documents))
	for i := range documents {
		out[i] = RerankScore{Index: i, Score: -float64(i)}
	}
	return out, nil
}

// isNoopReranker reports whether the configured reranker is the disabled/noop
// one, for logging and diagnostics.
func isNoopReranker(r RerankerPort) bool {
	_, ok := r.(NoopReranker)
	return ok
}

// HTTPReranker calls a Hugging Face text-embeddings-inference (TEI) /rerank
// endpoint, which serves cross-encoder models such as BAAI/bge-reranker-v2-m3.
type HTTPReranker struct {
	baseURL string
	client  *http.Client
}

// NewHTTPReranker builds a reranker against the given base URL. A nil client
// falls back to a short-timeout default so a wedged sidecar cannot stall chat.
func NewHTTPReranker(baseURL string, client *http.Client) *HTTPReranker {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	return &HTTPReranker{baseURL: strings.TrimRight(baseURL, "/"), client: client}
}

type teiRerankItem struct {
	Index int     `json:"index"`
	Score float64 `json:"score"`
}

func (h *HTTPReranker) Rerank(ctx context.Context, query string, documents []string) ([]RerankScore, error) {
	if h.baseURL == "" || query == "" || len(documents) == 0 {
		return nil, nil
	}
	body, err := json.Marshal(map[string]any{
		"query": query,
		"texts": documents,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal rerank request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.baseURL+"/rerank", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build rerank request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("rerank request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("rerank returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var raw []teiRerankItem
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode rerank response: %w", err)
	}
	out := make([]RerankScore, len(raw))
	for i, item := range raw {
		out[i] = RerankScore{Index: item.Index, Score: item.Score}
	}
	return out, nil
}

// rerankAndTrim reranks the retrieved chunks by cross-encoder score and returns
// the top limit. It only runs when there are more candidates than limit (no
// point trimming an already-small set, and it skips the network call entirely
// for short result sets). Any rerank failure falls back to the SQL order trimmed
// to limit, so chat never breaks on a reranker outage.
func rerankAndTrim(ctx context.Context, reranker RerankerPort, query string, chunks []repository.ChunkMatch, limit int) []repository.ChunkMatch {
	if len(chunks) <= limit {
		return chunks
	}
	if limit <= 0 {
		return chunks
	}

	documents := make([]string, len(chunks))
	for i, c := range chunks {
		documents[i] = c.ChunkText
	}

	scores, err := reranker.Rerank(ctx, query, documents)
	if err != nil || len(scores) == 0 {
		return chunks[:limit]
	}

	scoreByIndex := make(map[int]float64, len(scores))
	for _, s := range scores {
		if s.Index >= 0 && s.Index < len(chunks) {
			scoreByIndex[s.Index] = s.Score
		}
	}

	order := make([]int, len(chunks))
	for i := range chunks {
		order[i] = i
	}
	// Stable sort by rerank score desc; ties keep the original SQL order.
	sort.SliceStable(order, func(a, b int) bool {
		return scoreByIndex[order[a]] > scoreByIndex[order[b]]
	})

	trimmed := make([]repository.ChunkMatch, 0, limit)
	for i := 0; i < limit && i < len(order); i++ {
		trimmed = append(trimmed, chunks[order[i]])
	}
	return trimmed
}
