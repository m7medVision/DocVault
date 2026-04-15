package application

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/docvault/reminder/internal/reminder"
)

type stubDateExtractor struct {
	dates *reminder.ExtractedDates
	err   error
}

func (s stubDateExtractor) ExtractDates(ctx context.Context, text string, docType string) (*reminder.ExtractedDates, error) {
	return s.dates, s.err
}

type stubReminderStore struct {
	saved []*reminder.ReminderRule
}

func (s *stubReminderStore) SaveReminderRule(ctx context.Context, rule *reminder.ReminderRule) error {
	s.saved = append(s.saved, rule)
	return nil
}

func TestReminderJobHandlerSkipsWhenLLMExtractionFails(t *testing.T) {
	handler := NewReminderJobHandler(
		reminder.NewReminderExtractor([]int{30, 7, 1}),
		stubDateExtractor{err: context.DeadlineExceeded},
		&stubReminderStore{},
	)

	body, err := json.Marshal(QueueMessage{
		JobID:      "job-1",
		DocumentID: "doc-1",
		TenantID:   "tenant-1",
		SourceText: "Warranty End Date: April 14, 2027",
	})
	if err != nil {
		t.Fatalf("Marshal() unexpected error: %v", err)
	}

	err = handler.Handle(context.Background(), amqp.Delivery{Body: body})
	if err != nil {
		t.Fatalf("Handle() unexpected error: %v", err)
	}
}

func TestReminderJobHandlerCreatesRuleFromLLMExpiryDate(t *testing.T) {
	expiryDate := time.Now().AddDate(1, 0, 0)
	store := &stubReminderStore{}
	handler := NewReminderJobHandler(
		reminder.NewReminderExtractor([]int{30, 7, 1}),
		stubDateExtractor{dates: &reminder.ExtractedDates{
			ExpiryDate: &expiryDate,
			DatesFound: []reminder.DateMatch{{
				Date:       expiryDate,
				Pattern:    "llm.expiry_date",
				Context:    "Warranty End Date",
				Confidence: 0.99,
			}},
		}},
		store,
	)

	docType := "warranty"
	body, err := json.Marshal(QueueMessage{
		JobID:        "job-2",
		DocumentID:   "doc-2",
		TenantID:     "tenant-1",
		SourceText:   "Warranty End Date: April 14, 2027",
		DocumentType: &docType,
	})
	if err != nil {
		t.Fatalf("Marshal() unexpected error: %v", err)
	}

	err = handler.Handle(context.Background(), amqp.Delivery{Body: body})
	if err != nil {
		t.Fatalf("Handle() unexpected error: %v", err)
	}

	if len(store.saved) != 1 {
		t.Fatalf("saved rules = %d, want 1", len(store.saved))
	}
	if store.saved[0].RuleType != reminder.RuleTypeExpiry {
		t.Fatalf("rule type = %q, want %q", store.saved[0].RuleType, reminder.RuleTypeExpiry)
	}
}
