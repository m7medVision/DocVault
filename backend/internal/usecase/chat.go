package usecase

import (
	"bufio"
	"bytes"
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

const defaultChatBaseURL = "https://openrouter.ai/api/v1"

// minimumChatGroundingScore is the relevance floor for chunks fed to the chat
// model. It is deliberately higher than the search-wide minimumSearchScore so
// that weak, off-topic matches do not pollute the grounding context (and the
// "couldn't find anything" path triggers for genuinely irrelevant questions).
const minimumChatGroundingScore = 0.15

const systemPrompt = `You are a helpful assistant that answers questions strictly from the user's documents.
You are given a set of numbered passages retrieved from those documents.
Answer ONLY using the information in the numbered passages below. Do not invent facts.
Cite every supporting passage inline using its number in square brackets, for example [1] or [2].
If the answer cannot be found in the passages, say so plainly and do not guess.
Reply in the user's language (Arabic or English).
Use Markdown only when it improves readability. Do not restate the user's question.`

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatInput struct {
	DocumentID string
	Messages   []ChatMessage
	TenantID   string
	OrgID      string
	APIKey     string
	ChatModel  string
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
	embedder    search.Embedder
	searchRepo  repository.SearchRepository
	httpClient  *http.Client
	chatBaseURL string
}

func NewChatService(embedder search.Embedder, searchRepo repository.SearchRepository) *ChatService {
	return &ChatService{
		embedder:    embedder,
		searchRepo:  searchRepo,
		httpClient:  http.DefaultClient,
		chatBaseURL: defaultChatBaseURL,
	}
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

	embedding, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to generate query embedding: %w", err)
	}
	if len(embedding) != expectedEmbeddingDimensions {
		return fmt.Errorf("unexpected embedding size: got %d want %d", len(embedding), expectedEmbeddingDimensions)
	}

	searchResult, err := s.searchRepo.Search(ctx, repository.SearchRequest{
		Query:       query,
		TenantID:    input.TenantID,
		OrgID:       input.OrgID,
		QueryVector: search.FormatVectorLiteral(embedding),
		Limit:       10,
		MinScore:    minimumChatGroundingScore,
		DocumentID:  input.DocumentID,
	})
	if err != nil {
		return fmt.Errorf("retrieval failed: %w", err)
	}

	contextBlock, sources := buildRetrievalContext(searchResult.Chunks)

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

	apiMessages := []map[string]string{
		{"role": "system", "content": systemContent},
	}
	for _, msg := range input.Messages {
		apiMessages = append(apiMessages, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	reqBody := map[string]interface{}{
		"model":    input.ChatModel,
		"messages": apiMessages,
		"stream":   true,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := strings.TrimRight(s.chatBaseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+input.APIKey)

	client := s.httpClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		slog.Error("OpenRouter API error", "status", resp.StatusCode, "body", string(respBody))
		return fmt.Errorf("OpenRouter API error: status %d", resp.StatusCode)
	}

	writeChunk(map[string]interface{}{
		"type":      "TEXT_MESSAGE_START",
		"messageId": messageID,
		"role":      "assistant",
		"timestamp": timestamp,
	})

	scanner := bufio.NewScanner(resp.Body)
	hasContent := false
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			var chunk map[string]interface{}
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			choices, ok := chunk["choices"].([]interface{})
			if !ok || len(choices) == 0 {
				continue
			}

			choice, ok := choices[0].(map[string]interface{})
			if !ok {
				continue
			}

			delta, ok := choice["delta"].(map[string]interface{})
			if !ok {
				continue
			}

			content, ok := delta["content"].(string)
			if !ok || content == "" {
				continue
			}

			hasContent = true
			writeChunk(map[string]interface{}{
				"type":      "TEXT_MESSAGE_CONTENT",
				"messageId": messageID,
				"delta":     content,
				"timestamp": timestamp,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		slog.Error("error reading streaming response", "error", err)
		writeChunk(map[string]interface{}{
			"type":      "RUN_ERROR",
			"error":     map[string]string{"message": fmt.Sprintf("error reading streaming response: %v", err)},
			"timestamp": timestamp,
		})
		return fmt.Errorf("error reading streaming response: %w", err)
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
