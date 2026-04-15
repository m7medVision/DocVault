package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/docvault/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrReminderNotFound = errors.New("reminder not found")

// reminderRepository handles reminder data access.
type reminderRepository struct {
	db *pgxpool.Pool
}

// NewReminderRepository creates a new ReminderRepository.
func NewReminderRepository(db *pgxpool.Pool) ReminderRepository {
	return &reminderRepository{db: db}
}

// Create creates a new reminder rule.
func (r *reminderRepository) Create(ctx context.Context, reminder *model.ReminderRule) error {
	if reminder == nil {
		return fmt.Errorf("reminder is nil")
	}
	query := `
		INSERT INTO reminder_rules (id, document_id, tenant_id, rule_type, trigger_date, notify_days_before, source, active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
	`
	_, err := r.db.Exec(ctx, query,
		reminder.ID, reminder.DocumentID, reminder.TenantID,
		reminder.RuleType, reminder.TriggerDate, reminder.NotifyDaysBefore,
		reminder.Source, reminder.Active,
	)
	if err != nil {
		return fmt.Errorf("failed to create reminder: %w", err)
	}
	return nil
}

// GetByID retrieves a reminder by ID.
func (r *reminderRepository) GetByID(ctx context.Context, tenantID, id string) (*model.ReminderRule, error) {
	query := `
		SELECT id, document_id, tenant_id, rule_type, trigger_date, notify_days_before, source, active, created_at
		FROM reminder_rules
		WHERE id = $1 AND tenant_id = $2
	`
	var reminder model.ReminderRule
	err := r.db.QueryRow(ctx, query, id, tenantID).Scan(
		&reminder.ID, &reminder.DocumentID, &reminder.TenantID,
		&reminder.RuleType, &reminder.TriggerDate, &reminder.NotifyDaysBefore,
		&reminder.Source, &reminder.Active, &reminder.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrReminderNotFound
		}
		return nil, fmt.Errorf("failed to get reminder: %w", err)
	}
	return &reminder, nil
}

// GetByDocument retrieves reminders for a document.
func (r *reminderRepository) GetByDocument(ctx context.Context, tenantID, documentID string) ([]model.ReminderRule, error) {
	query := `
		SELECT id, document_id, tenant_id, rule_type, trigger_date, notify_days_before, source, active, created_at
		FROM reminder_rules
		WHERE document_id = $1 AND tenant_id = $2
		ORDER BY trigger_date ASC
	`
	rows, err := r.db.Query(ctx, query, documentID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reminders: %w", err)
	}
	defer rows.Close()

	reminders := make([]model.ReminderRule, 0)
	for rows.Next() {
		var r model.ReminderRule
		if err := rows.Scan(
			&r.ID, &r.DocumentID, &r.TenantID, &r.RuleType,
			&r.TriggerDate, &r.NotifyDaysBefore, &r.Source, &r.Active, &r.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan reminder: %w", err)
		}
		reminders = append(reminders, r)
	}
	return reminders, nil
}

// ListByTenant lists reminders for a tenant, optionally filtering to active only.
func (r *reminderRepository) ListByTenant(ctx context.Context, tenantID string, activeOnly bool) ([]model.ReminderRule, error) {
	query := `
		SELECT id, document_id, tenant_id, rule_type, trigger_date, notify_days_before, source, active, created_at
		FROM reminder_rules
		WHERE tenant_id = $1
	`
	args := []interface{}{tenantID}
	if activeOnly {
		query += ` AND active = true`
	}
	query += ` ORDER BY trigger_date ASC, created_at DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list reminders: %w", err)
	}
	defer rows.Close()

	reminders := make([]model.ReminderRule, 0)
	for rows.Next() {
		var reminder model.ReminderRule
		if err := rows.Scan(
			&reminder.ID, &reminder.DocumentID, &reminder.TenantID, &reminder.RuleType,
			&reminder.TriggerDate, &reminder.NotifyDaysBefore, &reminder.Source, &reminder.Active, &reminder.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan reminder: %w", err)
		}
		reminders = append(reminders, reminder)
	}

	return reminders, nil
}

// ListUpcoming lists reminders due within the given number of days.
func (r *reminderRepository) ListUpcoming(ctx context.Context, tenantID string, withinDays int) ([]model.ReminderRule, error) {
	query := `
		SELECT rr.id, rr.document_id, rr.tenant_id, rr.rule_type, rr.trigger_date, rr.notify_days_before, rr.source, rr.active, rr.created_at
		FROM reminder_rules rr
		WHERE rr.tenant_id = $1
		  AND rr.active = true
		  AND rr.trigger_date <= CURRENT_DATE + ($2::integer)
		ORDER BY rr.trigger_date ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, withinDays)
	if err != nil {
		return nil, fmt.Errorf("failed to list upcoming reminders: %w", err)
	}
	defer rows.Close()

	reminders := make([]model.ReminderRule, 0)
	for rows.Next() {
		var r model.ReminderRule
		if err := rows.Scan(
			&r.ID, &r.DocumentID, &r.TenantID, &r.RuleType,
			&r.TriggerDate, &r.NotifyDaysBefore, &r.Source, &r.Active, &r.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan reminder: %w", err)
		}
		reminders = append(reminders, r)
	}
	return reminders, nil
}

// Update updates a reminder rule.
func (r *reminderRepository) Update(ctx context.Context, reminder *model.ReminderRule) error {
	if reminder == nil {
		return fmt.Errorf("reminder is nil")
	}
	query := `
		UPDATE reminder_rules
		SET rule_type = $1, trigger_date = $2, notify_days_before = $3, active = $4
		WHERE id = $5 AND tenant_id = $6
	`
	_, err := r.db.Exec(ctx, query,
		reminder.RuleType, reminder.TriggerDate, reminder.NotifyDaysBefore,
		reminder.Active, reminder.ID, reminder.TenantID,
	)
	if err != nil {
		return fmt.Errorf("failed to update reminder: %w", err)
	}
	return nil
}

// Delete deletes a reminder rule.
func (r *reminderRepository) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM reminder_rules WHERE id = $1 AND tenant_id = $2`
	result, err := r.db.Exec(ctx, query, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete reminder: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrReminderNotFound
	}
	return nil
}

// CreateEvent creates a reminder event.
func (r *reminderRepository) CreateEvent(ctx context.Context, event *model.ReminderEvent) error {
	if event == nil {
		return fmt.Errorf("event is nil")
	}
	query := `
		INSERT INTO reminder_events (id, rule_id, scheduled_at, sent_at, channel, status, error_message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`
	_, err := r.db.Exec(ctx, query,
		event.ID, event.RuleID, event.ScheduledAt, event.SentAt,
		event.Channel, event.Status, event.ErrorMessage,
	)
	if err != nil {
		return fmt.Errorf("failed to create reminder event: %w", err)
	}
	return nil
}

// UpdateEvent updates a reminder event.
func (r *reminderRepository) UpdateEvent(ctx context.Context, event *model.ReminderEvent) error {
	if event == nil {
		return fmt.Errorf("event is nil")
	}
	query := `
		UPDATE reminder_events
		SET sent_at = $1, status = $2, error_message = $3
		WHERE id = $4
	`
	_, err := r.db.Exec(ctx, query,
		event.SentAt, event.Status, event.ErrorMessage, event.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update reminder event: %w", err)
	}
	return nil
}

// GetPendingEvents retrieves pending reminder events.
func (r *reminderRepository) GetPendingEvents(ctx context.Context, before time.Time) ([]model.ReminderEvent, error) {
	query := `
		SELECT id, rule_id, scheduled_at, sent_at, channel, status, error_message, created_at
		FROM reminder_events
		WHERE status = 'pending' AND scheduled_at <= $1
		ORDER BY scheduled_at ASC
	`
	rows, err := r.db.Query(ctx, query, before)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending events: %w", err)
	}
	defer rows.Close()

	var events []model.ReminderEvent
	for rows.Next() {
		var e model.ReminderEvent
		if err := rows.Scan(
			&e.ID, &e.RuleID, &e.ScheduledAt, &e.SentAt,
			&e.Channel, &e.Status, &e.ErrorMessage, &e.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan reminder event: %w", err)
		}
		events = append(events, e)
	}
	return events, nil
}
