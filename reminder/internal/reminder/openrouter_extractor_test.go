package reminder

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenRouterDateExtractorExtractsWarrantyDates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("Authorization header = %q, want Bearer test-key", got)
		}

		var request map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() unexpected error: %v", err)
		}
		if request["model"] != "openai/gpt-4.1-mini" {
			t.Fatalf("model = %v, want openai/gpt-4.1-mini", request["model"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices": [{
				"message": {
					"content": "{\"issue_date\":\"2026-04-14\",\"due_date\":\"2026-04-14\",\"expiry_date\":\"2027-04-14\",\"renewal_date\":\"\",\"matches\":[{\"field\":\"issue_date\",\"date\":\"2026-04-14\",\"label\":\"Issue Date\",\"excerpt\":\"Issue Date: April 14, 2026\",\"confidence\":0.98},{\"field\":\"expiry_date\",\"date\":\"2027-04-14\",\"label\":\"Warranty End Date\",\"excerpt\":\"Warranty End Date: April 14, 2027\",\"confidence\":0.99}]}"
				}
			}]
		}`))
	}))
	defer server.Close()

	extractor := NewOpenRouterDateExtractor("test-key", "openai/gpt-4.1-mini", 4000, server.Client())
	extractor.baseURL = server.URL

	dates, err := extractor.ExtractDates(context.Background(), "Warranty End Date: April 14, 2027", "warranty")
	if err != nil {
		t.Fatalf("ExtractDates() unexpected error: %v", err)
	}
	if dates.ExpiryDate == nil {
		t.Fatal("expected expiry date to be extracted")
	}
	if got := dates.ExpiryDate.Format("2006-01-02"); got != "2027-04-14" {
		t.Fatalf("expiry date = %s, want 2027-04-14", got)
	}
	if len(dates.DatesFound) != 2 {
		t.Fatalf("dates found = %d, want 2", len(dates.DatesFound))
	}
}

func TestOpenRouterDateExtractorRejectsInvalidDatePayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"issue_date\":\"14/04/2026\",\"due_date\":\"\",\"expiry_date\":\"\",\"renewal_date\":\"\",\"matches\":[]}"}}]}`))
	}))
	defer server.Close()

	extractor := NewOpenRouterDateExtractor("test-key", "openai/gpt-4.1-mini", 4000, server.Client())
	extractor.baseURL = server.URL

	_, err := extractor.ExtractDates(context.Background(), "Issue Date: April 14, 2026", "invoice")
	if err == nil {
		t.Fatal("ExtractDates() expected error for invalid date payload")
	}
}

func TestOpenRouterDateExtractorTruncatesLongInput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() unexpected error: %v", err)
		}
		if strings.Count(request.Messages[1].Content, "A") != 120 {
			t.Fatalf("truncated document text length = %d, want 120", strings.Count(request.Messages[1].Content, "A"))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"issue_date\":\"\",\"due_date\":\"\",\"expiry_date\":\"\",\"renewal_date\":\"\",\"matches\":[]}"}}]}`))
	}))
	defer server.Close()

	extractor := NewOpenRouterDateExtractor("test-key", "openai/gpt-4.1-mini", 120, server.Client())
	extractor.baseURL = server.URL

	_, err := extractor.ExtractDates(context.Background(), strings.Repeat("A", 500), "warranty")
	if err != nil {
		t.Fatalf("ExtractDates() unexpected error: %v", err)
	}
}

func TestOpenRouterDateExtractorFallbackOn429(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("Decode() unexpected error: %v", err)
		}

		callCount++
		model := request["model"].(string)

		if model == "openai/gpt-4.1-mini" {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":{"message":"rate limited"}}`))
			return
		}

		if model != "google/gemini-2.5-flash" {
			t.Fatalf("fallback model = %v, want google/gemini-2.5-flash", model)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"issue_date\":\"2026-04-14\",\"due_date\":\"\",\"expiry_date\":\"2027-04-14\",\"renewal_date\":\"\",\"matches\":[{\"field\":\"expiry_date\",\"date\":\"2027-04-14\",\"label\":\"Expiry\",\"excerpt\":\"Expires April 14, 2027\",\"confidence\":0.95}]}"}}]}`))
	}))
	defer server.Close()

	extractor := NewOpenRouterDateExtractor("test-key", "openai/gpt-4.1-mini", 4000, server.Client(), "google/gemini-2.5-flash")
	extractor.baseURL = server.URL

	dates, err := extractor.ExtractDates(context.Background(), "Expires April 14, 2027", "warranty")
	if err != nil {
		t.Fatalf("ExtractDates() unexpected error: %v", err)
	}
	if dates.ExpiryDate == nil {
		t.Fatal("expected expiry date to be extracted via fallback")
	}
	if callCount != 2 {
		t.Fatalf("call count = %d, want 2 (primary + fallback)", callCount)
	}
}

func TestOpenRouterDateExtractorNoFallbackWhenNotConfigured(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"message":"rate limited"}}`))
	}))
	defer server.Close()

	extractor := NewOpenRouterDateExtractor("test-key", "openai/gpt-4.1-mini", 4000, server.Client())
	extractor.baseURL = server.URL

	_, err := extractor.ExtractDates(context.Background(), "Expires April 14, 2027", "warranty")
	if err == nil {
		t.Fatal("ExtractDates() expected error when no fallback configured")
	}
	if !strings.Contains(err.Error(), "status 429") {
		t.Fatalf("error = %q, want status 429", err.Error())
	}
}

func TestOpenRouterDateExtractorFallbackAlsoFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error":{"message":"unavailable"}}`))
	}))
	defer server.Close()

	extractor := NewOpenRouterDateExtractor("test-key", "openai/gpt-4.1-mini", 4000, server.Client(), "google/gemini-2.5-flash")
	extractor.baseURL = server.URL

	_, err := extractor.ExtractDates(context.Background(), "Expires April 14, 2027", "warranty")
	if err == nil {
		t.Fatal("ExtractDates() expected error when both models fail")
	}
	if !strings.Contains(err.Error(), "status 503") {
		t.Fatalf("error = %q, want status 503", err.Error())
	}
}

func TestOpenRouterDateExtractorNoFallbackOnPermanentError(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"message":"bad request"}}`))
	}))
	defer server.Close()

	extractor := NewOpenRouterDateExtractor("test-key", "openai/gpt-4.1-mini", 4000, server.Client(), "google/gemini-2.5-flash")
	extractor.baseURL = server.URL

	_, err := extractor.ExtractDates(context.Background(), "Expires April 14, 2027", "warranty")
	if err == nil {
		t.Fatal("ExtractDates() expected error on permanent failure")
	}
	if callCount != 1 {
		t.Fatalf("call count = %d, want 1 (no fallback for permanent error)", callCount)
	}
}
