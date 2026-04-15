package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/docvault/backend/internal/model"
)

type AIService struct {
	apiKey string
	model  string
	client *http.Client
}

func NewAIService(apiKey, model string) *AIService {
	return &AIService{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type SuggestionInput struct {
	DocumentID    string
	Title         string
	DocType       string
	OCRText       string
	ExistingNames []string
}

type SuggestFolderOutput struct {
	SuggestedFolderID   *string `json:"suggested_folder_id,omitempty"`
	SuggestedFolderName string  `json:"suggested_folder_name,omitempty"`
	SuggestedName       string  `json:"suggested_name,omitempty"`
	Confidence          float32 `json:"confidence"`
	ShouldCreateNew     bool    `json:"should_create_new"`
}

func (s *AIService) SuggestFolder(ctx context.Context, tenantID string, folders []model.Folder, input *SuggestionInput) (*SuggestFolderOutput, error) {
	if s.apiKey == "" {
		return s.fallbackSuggestion(folders, input)
	}

	folderNames := make([]string, len(folders))
	folderPaths := make(map[string]string)
	for i, f := range folders {
		folderNames[i] = f.Name
		folderPaths[f.ID] = f.Name
	}

	prompt := buildSuggestionPrompt(folderNames, input)

	reqBody := map[string]interface{}{
		"model": s.model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a document organization assistant. Analyze the document information and suggest the best folder and name. Respond ONLY with valid JSON in this exact format: {\"suggested_folder_name\": \"...\", \"suggested_name\": \"...\", \"confidence\": 0.0-1.0, \"should_create_new\": true/false}"},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.3,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return s.fallbackSuggestion(folders, input)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return s.fallbackSuggestion(folders, input)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("HTTP-Referer", "https://docvault.app")
	req.Header.Set("X-Title", "DocVault")

	resp, err := s.client.Do(req)
	if err != nil {
		return s.fallbackSuggestion(folders, input)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return s.fallbackSuggestion(folders, input)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return s.fallbackSuggestion(folders, input)
	}

	if len(result.Choices) == 0 || result.Choices[0].Message.Content == "" {
		return s.fallbackSuggestion(folders, input)
	}

	return parseAIResponse(result.Choices[0].Message.Content, folders, input)
}

func buildSuggestionPrompt(folderNames []string, input *SuggestionInput) string {
	var sb strings.Builder
	sb.WriteString("Analyze this document and suggest the best folder and name.\n\n")
	sb.WriteString("Available folders: ")
	if len(folderNames) == 0 {
		sb.WriteString("(no folders exist yet)\n")
	} else {
		sb.WriteString(strings.Join(folderNames, ", "))
		sb.WriteString("\n")
	}
	sb.WriteString("Document title: ")
	sb.WriteString(input.Title)
	sb.WriteString("\nDocument type: ")
	sb.WriteString(input.DocType)
	sb.WriteString("\n")
	if input.OCRText != "" {
		text := input.OCRText
		if len(text) > 2000 {
			text = text[:2000] + "..."
		}
		sb.WriteString("Document content preview: ")
		sb.WriteString(text)
		sb.WriteString("\n")
	}
	sb.WriteString("\nRespond with JSON only.")
	return sb.String()
}

func parseAIResponse(content string, folders []model.Folder, input *SuggestionInput) (*SuggestFolderOutput, error) {
	content = strings.TrimSpace(content)

	var jsonStart int
	for i, c := range content {
		if c == '{' {
			jsonStart = i
			break
		}
	}

	if jsonStart > 0 {
		content = content[jsonStart:]
	}

	var parsed struct {
		SuggestedFolderName string  `json:"suggested_folder_name"`
		SuggestedName       string  `json:"suggested_name"`
		Confidence          float32 `json:"confidence"`
		ShouldCreateNew     bool    `json:"should_create_new"`
	}

	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return nil, err
	}

	output := &SuggestFolderOutput{
		SuggestedFolderName: parsed.SuggestedFolderName,
		SuggestedName:       parsed.SuggestedName,
		Confidence:          parsed.Confidence,
		ShouldCreateNew:     parsed.ShouldCreateNew,
	}

	if !parsed.ShouldCreateNew && parsed.SuggestedFolderName != "" {
		for _, f := range folders {
			if strings.EqualFold(f.Name, parsed.SuggestedFolderName) {
				output.SuggestedFolderID = &f.ID
				break
			}
		}
	}

	return output, nil
}

func (s *AIService) fallbackSuggestion(folders []model.Folder, input *SuggestionInput) (*SuggestFolderOutput, error) {
	output := &SuggestFolderOutput{
		Confidence:      0.5,
		ShouldCreateNew: false,
	}

	docTypeFolder := map[string]string{
		"invoice":  "Invoices",
		"contract": "Contracts",
		"identity": "Identity",
		"warranty": "Warranties",
		"receipt":  "Receipts",
	}

	if name, ok := docTypeFolder[input.DocType]; ok {
		output.SuggestedFolderName = name
		for _, f := range folders {
			if f.Name == name {
				output.SuggestedFolderID = &f.ID
				break
			}
		}
		if output.SuggestedFolderID == nil {
			output.ShouldCreateNew = true
		}
	}

	if output.SuggestedFolderID == nil && len(folders) > 0 {
		output.SuggestedFolderID = &folders[0].ID
		output.SuggestedFolderName = folders[0].Name
	}

	output.SuggestedName = input.Title

	return output, nil
}
