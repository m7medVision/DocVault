package service

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

	"github.com/docvault/backend/internal/repository"
)

const systemPrompt = `You are a helpful assistant that answers questions about a specific document. 
Use only the information provided in the document context below to answer the user's questions.
If the answer cannot be found in the document, say so clearly.
Be concise and accurate. When referencing specific parts of the document, mention the page number if available.`

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

type ChatService struct {
	documentRepo repository.DocumentRepository
}

func NewChatService(documentRepo repository.DocumentRepository) *ChatService {
	return &ChatService{documentRepo: documentRepo}
}

func (s *ChatService) StreamChat(ctx context.Context, input *ChatInput, w io.Writer) error {
	pages, err := s.documentRepo.GetPages(ctx, input.TenantID, input.DocumentID)
	if err != nil {
		return fmt.Errorf("failed to get document pages: %w", err)
	}

	var contextParts []string
	for _, page := range pages {
		if page.OCRText != nil && strings.TrimSpace(*page.OCRText) != "" {
			contextParts = append(contextParts, fmt.Sprintf("--- Page %d ---\n%s", page.PageNumber, *page.OCRText))
		}
	}

	if len(contextParts) == 0 {
		return fmt.Errorf("document has not been processed yet: no OCR text available")
	}

	documentContext := strings.Join(contextParts, "\n\n")

	systemContent := fmt.Sprintf("%s\n\n<Document>\n%s\n</Document>", systemPrompt, documentContext)

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

	req, err := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+input.APIKey)

	client := &http.Client{}
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

	flusher, canFlush := w.(http.Flusher)

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				fmt.Fprintf(w, "data: [DONE]\n\n")
				if canFlush {
					flusher.Flush()
				}
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

			sseChunk := map[string]interface{}{
				"choices": []map[string]interface{}{
					{
						"delta": map[string]string{
							"content": content,
						},
					},
				},
			}

			chunkBytes, err := json.Marshal(sseChunk)
			if err != nil {
				continue
			}

			fmt.Fprintf(w, "data: %s\n\n", string(chunkBytes))
			if canFlush {
				flusher.Flush()
			}
		}
	}

	if err := scanner.Err(); err != nil {
		slog.Error("error reading streaming response", "error", err)
		return fmt.Errorf("error reading streaming response: %w", err)
	}

	return nil
}
