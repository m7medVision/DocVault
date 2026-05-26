package reminder

import "time"

// ReminderRule defines when to send reminders for a document.
type ReminderRule struct {
	ID               string    `json:"id"`
	DocumentID       string    `json:"document_id"`
	TenantID         string    `json:"tenant_id"`
	RuleType         string    `json:"rule_type"`
	TriggerDate      time.Time `json:"trigger_date"`
	NotifyDaysBefore []int     `json:"notify_days_before"`
	Source           string    `json:"source"`
	Active           bool      `json:"active"`
	CreatedAt        time.Time `json:"created_at"`
}

// ReminderEventStatus represents notification delivery state.
type ReminderEventStatus string

const (
	ReminderEventStatusPending ReminderEventStatus = "pending"
	ReminderEventStatusSent    ReminderEventStatus = "sent"
	ReminderEventStatusFailed  ReminderEventStatus = "failed"
	ReminderEventStatusSnoozed ReminderEventStatus = "snoozed"
)

// ReminderEvent tracks notification delivery state.
type ReminderEvent struct {
	ID           string              `json:"id"`
	RuleID       string              `json:"rule_id"`
	ScheduledAt  time.Time           `json:"scheduled_at"`
	SentAt       *time.Time          `json:"sent_at,omitempty"`
	Channel      string              `json:"channel"`
	Status       ReminderEventStatus `json:"status"`
	ErrorMessage *string             `json:"error_message,omitempty"`
	CreatedAt    time.Time           `json:"created_at"`
}
