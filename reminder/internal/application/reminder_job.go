package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

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
	extractor     *reminder.ReminderExtractor
	dateExtractor reminder.DateExtractor
	db            ReminderRuleStore
	logger        *slog.Logger
}

type ReminderRuleStore interface {
	SaveReminderRule(ctx context.Context, rule *reminder.ReminderRule) error
}

func NewReminderJobHandler(extractor *reminder.ReminderExtractor, dateExtractor reminder.DateExtractor, db ReminderRuleStore) *ReminderJobHandler {
	return &ReminderJobHandler{
		extractor:     extractor,
		dateExtractor: dateExtractor,
		db:            db,
		logger:        slog.Default(),
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

	if h.dateExtractor == nil {
		h.logger.Info("reminder extraction disabled",
			"document_id", msg.DocumentID,
		)
		telemetry.GetProvider().RecordJob(ctx, "reminder", true, time.Since(start))
		return nil
	}

	dates, err := h.dateExtractor.ExtractDates(ctx, msg.SourceText, stringValue(msg.DocumentType))
	if err != nil {
		h.logger.Warn("llm date extraction failed",
			"document_id", msg.DocumentID,
			"error", err,
		)
		telemetry.RecordError(ctx, err)
		telemetry.GetProvider().RecordJob(ctx, "reminder", true, time.Since(start))
		return nil
	}

	h.logger.Info("dates extracted",
		"document_id", msg.DocumentID,
		"dates_found", len(dates.DatesFound),
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
			"reason", noReminderReason(dates),
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

func noReminderReason(dates *reminder.ExtractedDates) string {
	if dates == nil || len(dates.DatesFound) == 0 {
		return "no_dates_found"
	}
	return "no_future_dates_found"
}
