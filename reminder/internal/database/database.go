// Package database provides database operations for reminders.
package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/docvault/reminder/internal/reminder"
)

// ReminderStore handles reminder persistence.
type ReminderStore struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewReminderStore creates a new reminder store.
func NewReminderStore(ctx context.Context, databaseURL string) (*ReminderStore, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &ReminderStore{
		pool:   pool,
		logger: slog.Default(),
	}, nil
}

// SaveReminderRule persists a reminder rule.
func (s *ReminderStore) SaveReminderRule(ctx context.Context, rule *reminder.ReminderRule) error {
	query := `
		INSERT INTO reminder_rules (
			id, document_id, tenant_id, rule_type, trigger_date,
			notify_days_before, source, active, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			rule_type = EXCLUDED.rule_type,
			trigger_date = EXCLUDED.trigger_date,
			notify_days_before = EXCLUDED.notify_days_before,
			active = EXCLUDED.active
	`

	_, err := s.pool.Exec(ctx, query,
		rule.ID,
		rule.DocumentID,
		rule.TenantID,
		rule.RuleType,
		rule.TriggerDate,
		rule.NotifyDaysBefore,
		rule.Source,
		rule.Active,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to save reminder rule: %w", err)
	}

	s.logger.Info("reminder rule saved",
		"rule_id", rule.ID,
		"document_id", rule.DocumentID,
		"rule_type", rule.RuleType,
	)

	return nil
}

// CreateReminderEvent records a reminder event.
func (s *ReminderStore) CreateReminderEvent(ctx context.Context, event *reminder.ReminderEvent) error {
	query := `
		INSERT INTO reminder_events (
			id, rule_id, scheduled_at, sent_at, channel, status,
			error_message, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := s.pool.Exec(ctx, query,
		event.ID,
		event.RuleID,
		event.ScheduledAt,
		event.SentAt,
		event.Channel,
		event.Status,
		event.Error,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to create reminder event: %w", err)
	}

	return nil
}

// UpdateEventStatus updates the status and error message of an event.
func (s *ReminderStore) UpdateEventStatus(ctx context.Context, eventID, status, errorMsg string) error {
	query := `
		UPDATE reminder_events
		SET status = $1, error_message = $2
		WHERE id = $3
	`

	_, err := s.pool.Exec(ctx, query, status, errorMsg, eventID)
	if err != nil {
		return fmt.Errorf("failed to update event status: %w", err)
	}

	return nil
}

// MarkEventSent marks an event as sent with the sent_at timestamp.
func (s *ReminderStore) MarkEventSent(ctx context.Context, eventID string, sentAt *time.Time) error {
	query := `
		UPDATE reminder_events
		SET status = 'sent', sent_at = $1
		WHERE id = $2
	`

	_, err := s.pool.Exec(ctx, query, sentAt, eventID)
	if err != nil {
		return fmt.Errorf("failed to mark event sent: %w", err)
	}

	return nil
}

// GetActiveReminders gets all active reminders for a tenant.
func (s *ReminderStore) GetActiveReminders(ctx context.Context, tenantID string) ([]reminder.ReminderRule, error) {
	query := `
		SELECT id, document_id, tenant_id, rule_type, trigger_date,
			   notify_days_before, source, active, created_at
		FROM reminder_rules
		WHERE tenant_id = $1 AND active = true
		ORDER BY trigger_date ASC
	`

	rows, err := s.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query reminders: %w", err)
	}
	defer rows.Close()

	var rules []reminder.ReminderRule
	for rows.Next() {
		var rule reminder.ReminderRule
		err := rows.Scan(
			&rule.ID,
			&rule.DocumentID,
			&rule.TenantID,
			&rule.RuleType,
			&rule.TriggerDate,
			&rule.NotifyDaysBefore,
			&rule.Source,
			&rule.Active,
			&rule.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reminder: %w", err)
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

// GetAllActiveRules gets all active reminder rules.
func (s *ReminderStore) GetAllActiveRules(ctx context.Context) ([]reminder.ReminderRule, error) {
	query := `
		SELECT id, document_id, tenant_id, rule_type, trigger_date,
			   notify_days_before, source, active, created_at
		FROM reminder_rules
		WHERE active = true AND trigger_date > NOW()
		ORDER BY trigger_date ASC
	`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active rules: %w", err)
	}
	defer rows.Close()

	var rules []reminder.ReminderRule
	for rows.Next() {
		var rule reminder.ReminderRule
		err := rows.Scan(
			&rule.ID,
			&rule.DocumentID,
			&rule.TenantID,
			&rule.RuleType,
			&rule.TriggerDate,
			&rule.NotifyDaysBefore,
			&rule.Source,
			&rule.Active,
			&rule.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rule: %w", err)
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

// GetDocumentReminders gets all reminder rules for a document.
func (s *ReminderStore) GetDocumentReminders(ctx context.Context, documentID string) ([]reminder.ReminderRule, error) {
	query := `
		SELECT id, document_id, tenant_id, rule_type, trigger_date,
			   notify_days_before, source, active, created_at
		FROM reminder_rules
		WHERE document_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.pool.Query(ctx, query, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query document reminders: %w", err)
	}
	defer rows.Close()

	var rules []reminder.ReminderRule
	for rows.Next() {
		var rule reminder.ReminderRule
		err := rows.Scan(
			&rule.ID,
			&rule.DocumentID,
			&rule.TenantID,
			&rule.RuleType,
			&rule.TriggerDate,
			&rule.NotifyDaysBefore,
			&rule.Source,
			&rule.Active,
			&rule.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reminder: %w", err)
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

// DeactivateReminder deactivates a reminder rule.
func (s *ReminderStore) DeactivateReminder(ctx context.Context, ruleID string) error {
	query := `
		UPDATE reminder_rules
		SET active = false
		WHERE id = $1
	`

	result, err := s.pool.Exec(ctx, query, ruleID)
	if err != nil {
		return fmt.Errorf("failed to deactivate reminder: %w", err)
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// Close closes the database connection pool.
func (s *ReminderStore) Close() {
	s.pool.Close()
}

// InitSchema creates the necessary tables if they don't exist.
func (s *ReminderStore) InitSchema(ctx context.Context) error {
	schema := `
		DO $$ BEGIN
			CREATE TYPE reminder_rule_source AS ENUM ('auto', 'manual');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;

		DO $$ BEGIN
			CREATE TYPE reminder_event_status AS ENUM ('pending', 'sent', 'failed', 'snoozed');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;

		CREATE TABLE IF NOT EXISTS reminder_rules (
			id UUID PRIMARY KEY,
			document_id UUID NOT NULL,
			tenant_id UUID NOT NULL,
			rule_type VARCHAR(50) NOT NULL,
			trigger_date DATE NOT NULL,
			notify_days_before INTEGER[] NOT NULL,
			source reminder_rule_source NOT NULL DEFAULT 'auto',
			active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS reminder_events (
			id UUID PRIMARY KEY,
			rule_id UUID NOT NULL REFERENCES reminder_rules(id),
			scheduled_at TIMESTAMPTZ NOT NULL,
			sent_at TIMESTAMPTZ,
			channel VARCHAR(50) NOT NULL,
			status reminder_event_status NOT NULL,
			error_message TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS notifications (
			id VARCHAR(255) PRIMARY KEY,
			tenant_id VARCHAR(255) NOT NULL,
			user_id VARCHAR(255) NOT NULL,
			type VARCHAR(50) NOT NULL DEFAULT 'reminder',
			title VARCHAR(500) NOT NULL,
			body TEXT,
			link VARCHAR(1000),
			status VARCHAR(50) NOT NULL DEFAULT 'unread',
			metadata JSONB,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			read_at TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_reminder_rules_tenant ON reminder_rules (tenant_id);
		CREATE INDEX IF NOT EXISTS idx_reminder_rules_trigger ON reminder_rules (trigger_date);
		CREATE INDEX IF NOT EXISTS idx_reminder_events_rule ON reminder_events (rule_id);
		CREATE INDEX IF NOT EXISTS idx_reminder_events_status ON reminder_events (status);
		CREATE INDEX IF NOT EXISTS idx_notifications_tenant_user ON notifications (tenant_id, user_id);
		CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications (status);
		CREATE INDEX IF NOT EXISTS idx_notifications_created ON notifications (created_at DESC);
	`

	_, err := s.pool.Exec(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	s.logger.Info("database schema initialized")
	return nil
}

// Notification represents an in-app notification.
type Notification struct {
	ID        string
	TenantID  string
	UserID    string
	Type      string
	Title     string
	Body      string
	Link      *string
	Status    string
	Metadata  map[string]interface{}
	CreatedAt time.Time
	ReadAt    *time.Time
}

// CreateNotification creates a new notification.
func (s *ReminderStore) CreateNotification(ctx context.Context, n *Notification) error {
	query := `
		INSERT INTO notifications (
			id, tenant_id, user_id, type, title, body, link, status, metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := s.pool.Exec(ctx, query,
		n.ID,
		n.TenantID,
		n.UserID,
		n.Type,
		n.Title,
		n.Body,
		n.Link,
		n.Status,
		nil, // metadata - JSONB handling
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	s.logger.Info("notification created",
		"notification_id", n.ID,
		"user_id", n.UserID,
		"tenant_id", n.TenantID,
	)

	return nil
}

// ListNotifications lists notifications for a user.
func (s *ReminderStore) ListNotifications(ctx context.Context, tenantID, userID string, status string, limit int) ([]Notification, error) {
	query := `
		SELECT id, tenant_id, user_id, type, title, body, link, status, metadata, created_at, read_at
		FROM notifications
		WHERE tenant_id = $1 AND user_id = $2
	`

	args := []interface{}{tenantID, userID}

	if status != "" {
		query += " AND status = $3"
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC LIMIT $" + fmt.Sprintf("%d", 3+len(args))

	if limit <= 0 || limit > 100 {
		limit = 20
	}
	query += fmt.Sprintf(" LIMIT %d", limit)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		err := rows.Scan(
			&n.ID,
			&n.TenantID,
			&n.UserID,
			&n.Type,
			&n.Title,
			&n.Body,
			&n.Link,
			&n.Status,
			&n.Metadata,
			&n.CreatedAt,
			&n.ReadAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		notifications = append(notifications, n)
	}

	return notifications, nil
}

// MarkNotificationRead marks a notification as read.
func (s *ReminderStore) MarkNotificationRead(ctx context.Context, notificationID, tenantID, userID string) error {
	query := `
		UPDATE notifications
		SET status = 'read', read_at = $1
		WHERE id = $2 AND tenant_id = $3 AND user_id = $4
	`

	result, err := s.pool.Exec(ctx, query, time.Now(), notificationID, tenantID, userID)
	if err != nil {
		return fmt.Errorf("failed to mark notification read: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}
