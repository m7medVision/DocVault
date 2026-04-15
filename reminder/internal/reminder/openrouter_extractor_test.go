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
		if request["model"] != "mistralai/mistral-large" {
			t.Fatalf("model = %v, want mistralai/mistral-large", request["model"])
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

	extractor := NewOpenRouterDateExtractor("test-key", "mistralai/mistral-large", 4000, server.Client())
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

	extractor := NewOpenRouterDateExtractor("test-key", "mistralai/mistral-large", 4000, server.Client())
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

	extractor := NewOpenRouterDateExtractor("test-key", "mistralai/mistral-large", 120, server.Client())
	extractor.baseURL = server.URL

	_, err := extractor.ExtractDates(context.Background(), strings.Repeat("A", 500), "warranty")
	if err != nil {
		t.Fatalf("ExtractDates() unexpected error: %v", err)
	}
}
