package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/docvault/backend/internal/repository"
	"github.com/docvault/backend/internal/search"
)

const systemPrompt = `You are a helpful assistant that answers questions strictly from the user's documents.
You are given a set of numbered passages retrieved from those documents.
Answer ONLY using the information in the numbered passages below. Do not invent facts.
Cite every supporting passage inline using its number in square brackets, for example [1] or [2].
If the answer cannot be found in the passages, say so plainly and do not guess.
Reply in the user's language (Arabic or English).
Use Markdown only when it improves readability. Do not restate the user's question.`

// defaultChatRetrieveK is the number of chunks fetched for grounding when no
// explicit value is supplied. A wider pool than the previous hard-coded 10
// improves recall for short / proper-noun queries and gives a reranker room to
// work; precision is protected by the top-K cap, lexical scoring, and the
// "say if not found" system prompt.
const defaultChatRetrieveK = 40

// defaultChatContextK is how many chunks are actually sent to the generator
// after reranking the wider retrieve pool. Keeps the prompt focused and bounded.
const defaultChatContextK = 10

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatInput struct {
	DocumentID     string
	Messages       []ChatMessage
	TenantID       string
	OrgID          string
	UserID         string
	GroupIDs       []string
	IsAdmin        bool
	APIKey         string
	ChatModel      string
	RetrieveK      int
	ContextK       int
	RewriteQueries bool
}

// ChatSource describes a single retrieved passage surfaced as a citation.
type ChatSource struct {
	N          int     `json:"n"`
	DocumentID string  `json:"documentId"`
	Title      string  `json:"title"`
	Page       int     `json:"page"`
	Score      float64 `json:"score"`
}

type ChatService struct {
	embedder   search.Embedder
	searchRepo repository.SearchRepository
	llm        LLMChatPort
	reranker   RerankerPort
}

func NewChatService(embedder search.Embedder, searchRepo repository.SearchRepository) *ChatService {
	return &ChatService{
		embedder:   embedder,
		searchRepo: searchRepo,
		llm:        NewOpenRouterChatClient(defaultChatBaseURL, http.DefaultClient),
		reranker:   NoopReranker{},
	}
}

// WithReranker installs a cross-encoder reranker. When unset, chat uses the
// noop reranker (SQL retrieval order) so it works with no sidecar running.
func (s *ChatService) WithReranker(r RerankerPort) *ChatService {
	if r != nil {
		s.reranker = r
	}
	return s
}

// buildRetrievalContext turns retrieved chunks into a numbered context block and a
// parallel slice of citation sources. It is pure so it can be unit-tested in isolation.
func buildRetrievalContext(chunks []repository.ChunkMatch) (string, []ChatSource) {
	if len(chunks) == 0 {
		return "", nil
	}

	var builder strings.Builder
	sources := make([]ChatSource, 0, len(chunks))
	for i, chunk := range chunks {
		n := i + 1
		if i > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString(fmt.Sprintf("[%d] (source: %s, page %d)\n%s", n, chunk.DocumentTitle, chunk.PageNumber, chunk.ChunkText))
		sources = append(sources, ChatSource{
			N:          n,
			DocumentID: chunk.DocumentID,
			Title:      chunk.DocumentTitle,
			Page:       chunk.PageNumber,
			Score:      chunk.Score,
		})
	}

	return builder.String(), sources
}

func lastUserMessage(messages []ChatMessage) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			return messages[i].Content
		}
	}
	return ""
}

func (s *ChatService) StreamChat(ctx context.Context, input *ChatInput, w io.Writer) error {
	flusher, canFlush := w.(http.Flusher)

	messageID := fmt.Sprintf("msg_%d", time.Now().UnixNano())
	timestamp := time.Now().UnixMilli()

	writeChunk := func(chunk map[string]interface{}) {
		chunkBytes, err := json.Marshal(chunk)
		if err != nil {
			return
		}
		fmt.Fprintf(w, "data: %s\n\n", string(chunkBytes))
		if canFlush {
			flusher.Flush()
		}
	}

	query := strings.TrimSpace(lastUserMessage(input.Messages))
	if query == "" {
		return fmt.Errorf("no user message to answer")
	}

	// Reformulate the latest turn into a standalone retrieval query so pronouns
	// and references resolve (e.g. "what is its expiry?" -> the document name).
	// Skipped for single-turn chats; best-effort, falling back to the raw query.
	retrievalQuery := query
	if input.RewriteQueries {
		retrievalQuery = rewriteQuery(ctx, s.llm, input.ChatModel, input.APIKey, query, input.Messages)
		slog.Info("chat query rewritten", "raw", query, "rewritten", retrievalQuery)
	}

	embedding, err := s.embedder.Embed(ctx, retrievalQuery)
	if err != nil {
		return fmt.Errorf("failed to generate query embedding: %w", err)
	}
	if len(embedding) != expectedEmbeddingDimensions {
		return fmt.Errorf("unexpected embedding size: got %d want %d", len(embedding), expectedEmbeddingDimensions)
	}

	retrieveK := input.RetrieveK
	if retrieveK <= 0 {
		retrieveK = defaultChatRetrieveK
	}

	searchResult, err := s.searchRepo.Search(ctx, repository.SearchRequest{
		Query:       retrievalQuery,
		TenantID:    input.TenantID,
		OrgID:       input.OrgID,
		UserID:      input.UserID,
		GroupIDs:    input.GroupIDs,
		IsAdmin:     input.IsAdmin,
		QueryVector: search.FormatVectorLiteral(embedding),
		Limit:       retrieveK,
		// Use the same relevance floor as /search: chat must never be less able to
		// find content than search. Off-topic chunks are kept out by the keyword
		// scoring in search.sql, the top-K limit, and the "say if not found" prompt.
		MinScore:   minimumSearchScore,
		DocumentID: input.DocumentID,
	})
	if err != nil {
		return fmt.Errorf("retrieval failed: %w", err)
	}

	contextK := input.ContextK
	if contextK <= 0 {
		contextK = defaultChatContextK
	}

	// Rerank the wider retrieve pool with a cross-encoder (when configured) and
	// trim to the grounding budget sent to the generator. Noop/disabled or a
	// reranker outage falls back to the SQL order.
	chunks := rerankAndTrim(ctx, s.reranker, retrievalQuery, searchResult.Chunks, contextK)

	contextBlock, sources := buildRetrievalContext(chunks)

	// No grounding available: stream a graceful message and finish without sources.
	if len(sources) == 0 {
		writeChunk(map[string]interface{}{
			"type":      "TEXT_MESSAGE_START",
			"messageId": messageID,
			"role":      "assistant",
			"timestamp": timestamp,
		})
		writeChunk(map[string]interface{}{
			"type":      "TEXT_MESSAGE_CONTENT",
			"messageId": messageID,
			"delta":     "I couldn't find anything about that in your documents.",
			"timestamp": timestamp,
		})
		writeChunk(map[string]interface{}{
			"type":         "TEXT_MESSAGE_END",
			"messageId":    messageID,
			"finishReason": "stop",
			"timestamp":    timestamp,
		})
		writeChunk(map[string]interface{}{
			"type":      "RUN_FINISHED",
			"timestamp": timestamp,
		})
		return nil
	}

	systemContent := fmt.Sprintf("%s\n\nNumbered passages:\n%s", systemPrompt, contextBlock)

	messages := make([]LLMMessage, 0, len(input.Messages)+1)
	messages = append(messages, LLMMessage{Role: "system", Content: systemContent})
	for _, msg := range input.Messages {
		messages = append(messages, LLMMessage{Role: msg.Role, Content: msg.Content})
	}

	started := false
	hasContent := false
	streamErr := s.llm.StreamCompletion(ctx, LLMChatRequest{
		Model:    input.ChatModel,
		APIKey:   input.APIKey,
		Messages: messages,
	},
		func() {
			started = true
			writeChunk(map[string]interface{}{
				"type":      "TEXT_MESSAGE_START",
				"messageId": messageID,
				"role":      "assistant",
				"timestamp": timestamp,
			})
		},
		func(delta string) {
			hasContent = true
			writeChunk(map[string]interface{}{
				"type":      "TEXT_MESSAGE_CONTENT",
				"messageId": messageID,
				"delta":     delta,
				"timestamp": timestamp,
			})
		},
	)
	if streamErr != nil {
		// A failure before the stream started (e.g. a non-200 from the provider)
		// is returned to the caller without an AG-UI event, exactly as before. A
		// mid-stream failure — after TEXT_MESSAGE_START was emitted — surfaces a
		// RUN_ERROR event first, matching the old scanner-error path.
		if started {
			writeChunk(map[string]interface{}{
				"type":      "RUN_ERROR",
				"error":     map[string]string{"message": streamErr.Error()},
				"timestamp": timestamp,
			})
		}
		return streamErr
	}

	if hasContent {
		writeChunk(map[string]interface{}{
			"type":         "TEXT_MESSAGE_END",
			"messageId":    messageID,
			"finishReason": "stop",
			"timestamp":    timestamp,
		})
	}

	writeChunk(map[string]interface{}{
		"type":      "SOURCES",
		"messageId": messageID,
		"sources":   sources,
		"timestamp": timestamp,
	})

	writeChunk(map[string]interface{}{
		"type":      "RUN_FINISHED",
		"timestamp": timestamp,
	})

	return nil
}
