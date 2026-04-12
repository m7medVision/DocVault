package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/docvault/reminder/internal/database"
	"github.com/docvault/reminder/internal/reminder"
	"github.com/docvault/reminder/internal/sentry"
	"github.com/docvault/reminder/internal/telemetry"
)

type QueueMessage struct {
	JobID        string            `json:"job_id"`
	DocumentID   string            `json:"document_id"`
	TenantID     string            `json:"tenant_id"`
	OrgID        string            `json:"org_id"`
	SourceText   string            `json:"source_text"`
	DocumentType *string           `json:"document_type,omitempty"`
	ExpiryDate   *string           `json:"expiry_date,omitempty"`
	Issuer       *string           `json:"issuer,omitempty"`
	Priority     string            `json:"priority"`
	CreatedAt    string            `json:"created_at"`
	RetryCount   int               `json:"retry_count"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type ReminderJobHandler struct {
	extractor *reminder.ReminderExtractor
	db        *database.ReminderStore
	logger    *slog.Logger
}

func NewReminderJobHandler(extractor *reminder.ReminderExtractor, db *database.ReminderStore) *ReminderJobHandler {
	return &ReminderJobHandler{
		extractor: extractor,
		db:        db,
		logger:    slog.Default(),
	}
}

func (h *ReminderJobHandler) Handle(ctx context.Context, delivery amqp.Delivery) error {
	var msg QueueMessage
	if err := json.Unmarshal(delivery.Body, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	start := time.Now()

	sentry.AddTenant(msg.TenantID)

	ctx, span := telemetry.StartSpan(ctx, "process_reminder_job",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("job.id", msg.JobID),
			attribute.String("document.id", msg.DocumentID),
			attribute.String("tenant.id", msg.TenantID),
			attribute.String("job.type", "reminder"),
		),
	)
	defer span.End()

	h.logger.Info("processing reminder job",
		"job_id", msg.JobID,
		"document_id", msg.DocumentID,
		"tenant_id", msg.TenantID,
	)

	dateEx := reminder.NewDateExtractor()
	parsedDates, err := dateEx.ExtractDates(msg.SourceText, "")
	if err != nil {
		h.logger.Warn("date extraction failed", "error", err)
	}

	dates := convertToExtractedDates(parsedDates)
	if dates.ExpiryDate == nil && msg.ExpiryDate != nil {
		if parsedExpiry, ok := parseMessageDate(*msg.ExpiryDate); ok {
			dates.ExpiryDate = &parsedExpiry
			dates.DatesFound = append(dates.DatesFound, reminder.DateMatch{
				Date:       parsedExpiry,
				Pattern:    "queue.expiry_date",
				Context:    "processing metadata",
				Confidence: 0.95,
			})
		}
	}

	h.logger.Info("dates extracted",
		"document_id", msg.DocumentID,
		"dates_found", len(parsedDates),
	)

	rules, err := h.extractor.ExtractRules(
		ctx,
		msg.DocumentID,
		msg.TenantID,
		dates,
		stringValue(msg.DocumentType),
	)
	if err != nil {
		telemetry.RecordError(ctx, err)
		telemetry.GetProvider().RecordJob(ctx, "reminder", false, time.Since(start))
		return fmt.Errorf("failed to extract reminder rules: %w", err)
	}

	if len(rules) == 0 {
		h.logger.Info("no reminder rules created",
			"document_id", msg.DocumentID,
			"reason", "no_future_dates_found",
		)
		telemetry.GetProvider().RecordJob(ctx, "reminder", true, time.Since(start))
		return nil
	}

	for i := range rules {
		rule := &rules[i]
		if err := h.db.SaveReminderRule(ctx, rule); err != nil {
			telemetry.RecordError(ctx, err)
			telemetry.GetProvider().RecordJob(ctx, "reminder", false, time.Since(start))
			return fmt.Errorf("failed to save reminder rule: %w", err)
		}

		h.logger.Info("reminder rule created",
			"rule_id", rule.ID,
			"document_id", rule.DocumentID,
			"rule_type", rule.RuleType,
			"trigger_date", rule.TriggerDate,
		)
	}

	telemetry.GetProvider().RecordJob(ctx, "reminder", true, time.Since(start))
	return nil
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func parseMessageDate(raw string) (time.Time, bool) {
	formats := []string{
		time.RFC3339,
		"2006-01-02",
		"02/01/2006",
		"01/02/2006",
		"02-01-2006",
		"01-02-2006",
	}

	raw = strings.TrimSpace(raw)
	for _, format := range formats {
		parsed, err := time.Parse(format, raw)
		if err == nil {
			return parsed, true
		}
	}

	return time.Time{}, false
}

func convertToExtractedDates(parsed []reminder.ParsedDate) *reminder.ExtractedDates {
	result := &reminder.ExtractedDates{}
	for _, d := range parsed {
		dm := reminder.DateMatch{
			Date:       d.Date,
			Pattern:    d.Pattern,
			Context:    d.Context,
			Confidence: d.Confidence,
		}
		result.DatesFound = append(result.DatesFound, dm)
		if strings.Contains(d.Pattern, "expir") || strings.Contains(d.Pattern, "valid") {
			result.ExpiryDate = &d.Date
		} else if strings.Contains(d.Pattern, "issue") || strings.Contains(d.Pattern, "dated") {
			result.IssueDate = &d.Date
		} else if strings.Contains(d.Pattern, "due") {
			result.DueDate = &d.Date
		} else if strings.Contains(d.Pattern, "renew") {
			result.RenewalDate = &d.Date
		}
	}
	return result
}
