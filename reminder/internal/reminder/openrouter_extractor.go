package reminder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultOpenRouterBaseURL             = "https://openrouter.ai/api/v1"
	defaultReminderExtractionMaxChars    = 16000
	defaultReminderExtractionHTTPTimeout = 45 * time.Second
	isoDateLayout                        = "2006-01-02"
)

type OpenRouterDateExtractor struct {
	apiKey        string
	model         string
	fallbackModel string
	maxChars      int
	baseURL       string
	client        *http.Client
}

type openRouterChatRequest struct {
	Model          string                   `json:"model"`
	Messages       []openRouterChatMessage  `json:"messages"`
	ResponseFormat openRouterResponseFormat `json:"response_format"`
	Plugins        []openRouterPlugin       `json:"plugins,omitempty"`
	MaxTokens      int                      `json:"max_tokens"`
	Temperature    float64                  `json:"temperature"`
}

type openRouterChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterResponseFormat struct {
	Type       string               `json:"type"`
	JSONSchema openRouterJSONSchema `json:"json_schema"`
}

type openRouterJSONSchema struct {
	Name   string                 `json:"name"`
	Strict bool                   `json:"strict"`
	Schema map[string]interface{} `json:"schema"`
}

type openRouterPlugin struct {
	ID string `json:"id"`
}

type openRouterChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type llmExtractionResult struct {
	IssueDate   string             `json:"issue_date"`
	DueDate     string             `json:"due_date"`
	ExpiryDate  string             `json:"expiry_date"`
	RenewalDate string             `json:"renewal_date"`
	Matches     []llmExtractionHit `json:"matches"`
}

type llmExtractionHit struct {
	Field      string  `json:"field"`
	Date       string  `json:"date"`
	Label      string  `json:"label"`
	Excerpt    string  `json:"excerpt"`
	Confidence float64 `json:"confidence"`
}

func NewOpenRouterDateExtractor(apiKey, model string, maxChars int, client *http.Client, fallbackModels ...string) *OpenRouterDateExtractor {
	if client == nil {
		client = &http.Client{Timeout: defaultReminderExtractionHTTPTimeout}
	}
	if maxChars <= 0 {
		maxChars = defaultReminderExtractionMaxChars
	}

	var fallback string
	if len(fallbackModels) > 0 {
		fallback = strings.TrimSpace(fallbackModels[0])
	}

	return &OpenRouterDateExtractor{
		apiKey:        strings.TrimSpace(apiKey),
		model:         strings.TrimSpace(model),
		fallbackModel: fallback,
		maxChars:      maxChars,
		baseURL:       defaultOpenRouterBaseURL,
		client:        client,
	}
}

func (e *OpenRouterDateExtractor) ExtractDates(ctx context.Context, text string, docType string) (*ExtractedDates, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return &ExtractedDates{}, nil
	}
	if e.apiKey == "" {
		return nil, fmt.Errorf("OPENROUTER_API_KEY is required for reminder extraction")
	}
	if e.model == "" {
		return nil, fmt.Errorf("REMINDER_EXTRACTION_MODEL is required for reminder extraction")
	}

	result, err := e.callModel(ctx, e.model, docType, trimmed)
	if err != nil {
		if isTransientHTTPError(err) && e.fallbackModel != "" {
			result, fallbackErr := e.callModel(ctx, e.fallbackModel, docType, trimmed)
			if fallbackErr == nil {
				return result, nil
			}
		}
		return nil, err
	}

	return result, nil
}

func (e *OpenRouterDateExtractor) callModel(ctx context.Context, model string, docType string, text string) (*ExtractedDates, error) {
	body, err := json.Marshal(openRouterChatRequest{
		Model: model,
		Messages: []openRouterChatMessage{
			{
				Role:    "system",
				Content: "You extract reminder-relevant dates from OCR document text. Return only valid JSON matching the provided schema. Do not invent dates. Prefer explicit dates. You may derive an expiry or renewal date only when the document clearly states both a start date and a duration.",
			},
			{
				Role:    "user",
				Content: e.buildPrompt(docType, text),
			},
		},
		ResponseFormat: openRouterResponseFormat{
			Type: "json_schema",
			JSONSchema: openRouterJSONSchema{
				Name:   "reminder_dates",
				Strict: true,
				Schema: reminderDateSchema(),
			},
		},
		Plugins:     []openRouterPlugin{{ID: "response-healing"}},
		MaxTokens:   800,
		Temperature: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal OpenRouter request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenRouter request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenRouter: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("OpenRouter API error: status %d body %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var response openRouterChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode OpenRouter response: %w", err)
	}
	if len(response.Choices) == 0 || strings.TrimSpace(response.Choices[0].Message.Content) == "" {
		return nil, fmt.Errorf("no reminder extraction returned from OpenRouter")
	}

	var parsed llmExtractionResult
	content := cleanStructuredJSON(response.Choices[0].Message.Content)
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return nil, fmt.Errorf("failed to decode reminder extraction payload: %w", err)
	}

	return parsed.toExtractedDates()
}

func (e *OpenRouterDateExtractor) buildPrompt(docType string, text string) string {
	return fmt.Sprintf(
		"Document type: %s\n\nExtract reminder-relevant dates from the following document text. Return dates as YYYY-MM-DD. Use an empty string when a field is absent. Include only dates that are explicit in the text, except you may derive an expiry or renewal date when the text clearly includes a start date and duration. Include short evidence in matches.\n\nDocument text:\n%s",
		strings.TrimSpace(docType),
		truncateRunes(text, e.maxChars),
	)
}

func reminderDateSchema() map[string]interface{} {
	dateField := map[string]interface{}{
		"type":        "string",
		"description": "ISO date in YYYY-MM-DD format or empty string when unavailable.",
	}

	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"issue_date":   dateField,
			"due_date":     dateField,
			"expiry_date":  dateField,
			"renewal_date": dateField,
			"matches": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"field": map[string]interface{}{
							"type": "string",
							"enum": []string{"issue_date", "due_date", "expiry_date", "renewal_date"},
						},
						"date": dateField,
						"label": map[string]interface{}{
							"type": "string",
						},
						"excerpt": map[string]interface{}{
							"type": "string",
						},
						"confidence": map[string]interface{}{
							"type": "number",
						},
					},
					"required":             []string{"field", "date", "label", "excerpt", "confidence"},
					"additionalProperties": false,
				},
			},
		},
		"required":             []string{"issue_date", "due_date", "expiry_date", "renewal_date", "matches"},
		"additionalProperties": false,
	}
}

func (r llmExtractionResult) toExtractedDates() (*ExtractedDates, error) {
	result := &ExtractedDates{}

	if parsed, ok, err := parseLLMDate(r.IssueDate); err != nil {
		return nil, fmt.Errorf("invalid issue_date: %w", err)
	} else if ok {
		result.IssueDate = &parsed
	}
	if parsed, ok, err := parseLLMDate(r.DueDate); err != nil {
		return nil, fmt.Errorf("invalid due_date: %w", err)
	} else if ok {
		result.DueDate = &parsed
	}
	if parsed, ok, err := parseLLMDate(r.ExpiryDate); err != nil {
		return nil, fmt.Errorf("invalid expiry_date: %w", err)
	} else if ok {
		result.ExpiryDate = &parsed
	}
	if parsed, ok, err := parseLLMDate(r.RenewalDate); err != nil {
		return nil, fmt.Errorf("invalid renewal_date: %w", err)
	} else if ok {
		result.RenewalDate = &parsed
	}

	for _, match := range r.Matches {
		parsed, ok, err := parseLLMDate(match.Date)
		if err != nil {
			return nil, fmt.Errorf("invalid match date for %s: %w", match.Field, err)
		}
		if !ok {
			continue
		}

		result.DatesFound = append(result.DatesFound, DateMatch{
			Date:       parsed,
			Pattern:    "llm." + match.Field,
			Context:    strings.TrimSpace(strings.TrimSpace(match.Label) + ": " + strings.TrimSpace(match.Excerpt)),
			Confidence: match.Confidence,
		})

		switch match.Field {
		case "issue_date":
			if result.IssueDate == nil {
				result.IssueDate = &parsed
			}
		case "due_date":
			if result.DueDate == nil {
				result.DueDate = &parsed
			}
		case "expiry_date":
			if result.ExpiryDate == nil {
				result.ExpiryDate = &parsed
			}
		case "renewal_date":
			if result.RenewalDate == nil {
				result.RenewalDate = &parsed
			}
		}
	}

	result.ensureTopLevelDateMatches()
	return result, nil
}

func (d *ExtractedDates) ensureTopLevelDateMatches() {
	if len(d.DatesFound) > 0 {
		return
	}
	appendIfPresent := func(field string, value *time.Time) {
		if value == nil {
			return
		}
		d.DatesFound = append(d.DatesFound, DateMatch{
			Date:       *value,
			Pattern:    "llm." + field,
			Context:    field,
			Confidence: 0.9,
		})
	}
	appendIfPresent("issue_date", d.IssueDate)
	appendIfPresent("due_date", d.DueDate)
	appendIfPresent("expiry_date", d.ExpiryDate)
	appendIfPresent("renewal_date", d.RenewalDate)
}

func parseLLMDate(raw string) (time.Time, bool, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return time.Time{}, false, nil
	}

	for _, layout := range []string{isoDateLayout, time.RFC3339, time.RFC3339Nano} {
		parsed, err := time.Parse(layout, trimmed)
		if err == nil {
			return parsed, true, nil
		}
	}

	return time.Time{}, false, fmt.Errorf("expected ISO date, got %q", raw)
}

func cleanStructuredJSON(content string) string {
	trimmed := strings.TrimSpace(content)
	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	return strings.TrimSpace(trimmed)
}

func truncateRunes(text string, maxChars int) string {
	if maxChars <= 0 {
		return text
	}
	runes := []rune(text)
	if len(runes) <= maxChars {
		return text
	}
	return string(runes[:maxChars])
}

func isTransientHTTPError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	for _, marker := range []string{"status 429", "status 500", "status 502", "status 503"} {
		if strings.Contains(msg, marker) {
			return true
		}
	}
	return false
}
