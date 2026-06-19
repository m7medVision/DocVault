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
)

const defaultChatBaseURL = "https://openrouter.ai/api/v1"

// LLMMessage is a single chat message sent to the provider.
type LLMMessage struct {
	Role    string
	Content string
}

// LLMChatRequest is a streaming chat-completion request.
type LLMChatRequest struct {
	Model    string
	APIKey   string
	Messages []LLMMessage
}

// LLMChatPort streams a chat completion from the upstream LLM provider. The
// provider's HTTP transport and its SSE wire format are the adapter's concern;
// the chat usecase keeps the retrieval and AG-UI event protocol. onStart fires
// exactly once after the provider accepts the request (HTTP 200) and before any
// delta; onDelta fires once per content delta in arrival order. An error
// returned before onStart means the request never began streaming.
type LLMChatPort interface {
	StreamCompletion(ctx context.Context, req LLMChatRequest, onStart func(), onDelta func(delta string)) error
}

// OpenRouterChatClient is the OpenRouter implementation of LLMChatPort. It
// speaks the OpenAI-compatible /chat/completions streaming SSE protocol.
type OpenRouterChatClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewOpenRouterChatClient builds a client against the given base URL. A nil
// httpClient falls back to http.DefaultClient and an empty baseURL falls back
// to the public OpenRouter endpoint.
func NewOpenRouterChatClient(baseURL string, client *http.Client) *OpenRouterChatClient {
	if client == nil {
		client = http.DefaultClient
	}
	if baseURL == "" {
		baseURL = defaultChatBaseURL
	}
	return &OpenRouterChatClient{httpClient: client, baseURL: baseURL}
}

// StreamCompletion sends the chat request and streams content deltas. It builds
// the OpenRouter request body, posts it, and parses the SSE response, invoking
// onStart after the 200 response and onDelta for each non-empty content delta.
func (c *OpenRouterChatClient) StreamCompletion(ctx context.Context, req LLMChatRequest, onStart func(), onDelta func(delta string)) error {
	apiMessages := make([]map[string]string, 0, len(req.Messages))
	for _, msg := range req.Messages {
		apiMessages = append(apiMessages, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	reqBody := map[string]interface{}{
		"model":    req.Model,
		"messages": apiMessages,
		"stream":   true,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := strings.TrimRight(c.baseURL, "/") + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		slog.Error("OpenRouter API error", "status", resp.StatusCode, "body", string(respBody))
		return fmt.Errorf("OpenRouter API error: status %d", resp.StatusCode)
	}

	if onStart != nil {
		onStart()
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
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

		onDelta(content)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading streaming response: %w", err)
	}
	return nil
}
