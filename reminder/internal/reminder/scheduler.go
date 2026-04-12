// Package reminder provides reminder scheduling and event management.
package reminder

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Scheduler handles scheduling of reminder events.
type Scheduler struct {
	store        SchedulerStore
	logger       *slog.Logger
	notifyAtTime time.Time
}

// SchedulerStore interface for database operations needed by scheduler.
type SchedulerStore interface {
	GetAllActiveRules(ctx context.Context) ([]ReminderRule, error)
	CreateReminderEvent(ctx context.Context, event *ReminderEvent) error
}

// NewScheduler creates a new reminder scheduler.
func NewScheduler(store SchedulerStore, logger *slog.Logger, notifyAtTime string) *Scheduler {
	// Parse time in HH:MM format
	t, err := time.Parse("15:04", notifyAtTime)
	if err != nil {
		t = time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC) // Default 09:00
	}

	return &Scheduler{
		store:        store,
		logger:       logger,
		notifyAtTime: t,
	}
}

// RunDaily runs the daily scheduler check.
func (s *Scheduler) RunDaily(ctx context.Context) error {
	return s.ProcessDueReminders(ctx, RealClock{})
}

// ProcessDueReminders processes reminders that are due using the provided clock.
func (s *Scheduler) ProcessDueReminders(ctx context.Context, clock Clock) error {
	s.logger.Info("running daily reminder scheduler")

	rules, err := s.store.GetAllActiveRules(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active rules: %w", err)
	}

	s.logger.Info("found active rules", "count", len(rules))

	today := clock.Now().Truncate(24 * time.Hour)

	for _, rule := range rules {
		if err := s.processRule(ctx, &rule, today); err != nil {
			s.logger.Error("failed to process rule",
				"rule_id", rule.ID,
				"error", err,
			)
			continue
		}
	}

	s.logger.Info("daily scheduler completed")
	return nil
}

// processRule processes a single rule and creates events for upcoming notifications.
func (s *Scheduler) processRule(ctx context.Context, rule *ReminderRule, today time.Time) error {
	triggerDate := rule.TriggerDate.Truncate(24 * time.Hour)

	// Check each notify day
	for _, daysBefore := range rule.NotifyDaysBefore {
		notifyDate := triggerDate.AddDate(0, 0, -daysBefore)

		// Check if we should create an event for this date
		if notifyDate.Before(today) {
			continue // Already passed
		}

		if notifyDate.After(today.Add(24 * time.Hour)) {
			continue // Too far in the future
		}

		// Create the event
		event := &ReminderEvent{
			ID:          generateEventID(),
			RuleID:      rule.ID,
			ScheduledAt: s.getScheduledTime(notifyDate),
			Channel:     "email", // Default channel
			Status:      EventStatusPending,
		}

		if err := s.store.CreateReminderEvent(ctx, event); err != nil {
			return fmt.Errorf("failed to create event: %w", err)
		}

		s.logger.Info("created reminder event",
			"event_id", event.ID,
			"rule_id", rule.ID,
			"notify_date", notifyDate.Format("2006-01-02"),
		)
	}

	return nil
}

// getScheduledTime returns the scheduled time for a notification.
func (s *Scheduler) getScheduledTime(date time.Time) time.Time {
	return time.Date(
		date.Year(), date.Month(), date.Day(),
		s.notifyAtTime.Hour(), s.notifyAtTime.Minute(), 0, 0,
		time.UTC,
	)
}

// ReminderEvent represents a reminder notification event.
type ReminderEvent struct {
	ID          string
	RuleID      string
	ScheduledAt time.Time
	SentAt      *time.Time
	Channel     string
	Status      string
	Error       string
}

const (
	EventStatusPending = "pending"
	EventStatusSent    = "sent"
	EventStatusFailed  = "failed"
	EventStatusSnoozed = "snoozed"
)

// generateEventID creates a unique event ID.
func generateEventID() string {
	return newUUIDString()
}
