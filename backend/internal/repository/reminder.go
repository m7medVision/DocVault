package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	sqldb "github.com/docvault/backend/internal/db"
	model "github.com/docvault/backend/internal/domain/reminder"
	"github.com/docvault/backend/internal/platform/pgconv"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrReminderNotFound = errors.New("reminder not found")

// reminderRepository handles reminder data access.
type reminderRepository struct {
	queries sqldb.Querier
}

// NewReminderRepository creates a new ReminderRepository.
func NewReminderRepository(db *pgxpool.Pool) ReminderRepository {
	return &reminderRepository{queries: sqldb.New(db)}
}

// Create creates a new reminder rule.
func (r *reminderRepository) Create(ctx context.Context, reminder *model.ReminderRule) error {
	if reminder == nil {
		return fmt.Errorf("reminder is nil")
	}
	err := r.queries.CreateReminderRule(ctx, sqldb.CreateReminderRuleParams{
		ID:               reminder.ID,
		DocumentID:       reminder.DocumentID,
		TenantID:         reminder.TenantID,
		RuleType:         reminder.RuleType,
		TriggerDate:      pgconv.DateFromTime(reminder.TriggerDate),
		NotifyDaysBefore: pgconv.Int32sFromInts(reminder.NotifyDaysBefore),
		Source:           sqldb.ReminderRuleSource(reminder.Source),
		Active:           reminder.Active,
	})
	if err != nil {
		return fmt.Errorf("failed to create reminder: %w", err)
	}
	return nil
}

// GetByID retrieves a reminder by ID.
func (r *reminderRepository) GetByID(ctx context.Context, tenantID, id string) (*model.ReminderRule, error) {
	reminder, err := r.queries.GetReminderRuleByID(ctx, sqldb.GetReminderRuleByIDParams{ID: id, TenantID: tenantID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrReminderNotFound
		}
		return nil, fmt.Errorf("failed to get reminder: %w", err)
	}
	modelReminder := toModelReminderRule(reminder)
	return &modelReminder, nil
}

// GetByDocument retrieves reminders for a document.
func (r *reminderRepository) GetByDocument(ctx context.Context, tenantID, documentID string) ([]model.ReminderRule, error) {
	reminders, err := r.queries.GetReminderRulesByDocument(ctx, sqldb.GetReminderRulesByDocumentParams{
		DocumentID: documentID,
		TenantID:   tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get reminders: %w", err)
	}
	return toModelReminderRules(reminders), nil
}

// ListByTenant lists reminders for a tenant, optionally filtering to active only.
func (r *reminderRepository) ListByTenant(ctx context.Context, tenantID string, activeOnly bool) ([]model.ReminderRule, error) {
	var (
		reminders []sqldb.ReminderRule
		err       error
	)
	if activeOnly {
		reminders, err = r.queries.ListActiveReminderRulesByTenant(ctx, sqldb.ListActiveReminderRulesByTenantParams{TenantID: tenantID})
	} else {
		reminders, err = r.queries.ListReminderRulesByTenant(ctx, sqldb.ListReminderRulesByTenantParams{TenantID: tenantID})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list reminders: %w", err)
	}
	return toModelReminderRules(reminders), nil
}

// ListUpcoming lists reminders due within the given number of days.
func (r *reminderRepository) ListUpcoming(ctx context.Context, tenantID string, withinDays int) ([]model.ReminderRule, error) {
	reminders, err := r.queries.ListUpcomingReminderRules(ctx, sqldb.ListUpcomingReminderRulesParams{
		TenantID:   tenantID,
		WithinDays: int32(withinDays),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list upcoming reminders: %w", err)
	}
	return toModelReminderRules(reminders), nil
}

// Update updates a reminder rule.
func (r *reminderRepository) Update(ctx context.Context, reminder *model.ReminderRule) error {
	if reminder == nil {
		return fmt.Errorf("reminder is nil")
	}
	err := r.queries.UpdateReminderRule(ctx, sqldb.UpdateReminderRuleParams{
		RuleType:         reminder.RuleType,
		TriggerDate:      pgconv.DateFromTime(reminder.TriggerDate),
		NotifyDaysBefore: pgconv.Int32sFromInts(reminder.NotifyDaysBefore),
		Active:           reminder.Active,
		ID:               reminder.ID,
		TenantID:         reminder.TenantID,
	})
	if err != nil {
		return fmt.Errorf("failed to update reminder: %w", err)
	}
	return nil
}

// Delete deletes a reminder rule.
func (r *reminderRepository) Delete(ctx context.Context, tenantID, id string) error {
	rowsAffected, err := r.queries.DeleteReminderRule(ctx, sqldb.DeleteReminderRuleParams{ID: id, TenantID: tenantID})
	if err != nil {
		return fmt.Errorf("failed to delete reminder: %w", err)
	}
	if rowsAffected == 0 {
		return ErrReminderNotFound
	}
	return nil
}

// CreateEvent creates a reminder event.
func (r *reminderRepository) CreateEvent(ctx context.Context, event *model.ReminderEvent) error {
	if event == nil {
		return fmt.Errorf("event is nil")
	}
	err := r.queries.CreateReminderEvent(ctx, sqldb.CreateReminderEventParams{
		ID:           event.ID,
		RuleID:       event.RuleID,
		ScheduledAt:  pgconv.TimestamptzFromTime(event.ScheduledAt),
		SentAt:       pgconv.TimestamptzFromTimePtr(event.SentAt),
		Channel:      event.Channel,
		Status:       sqldb.ReminderEventStatus(event.Status),
		ErrorMessage: event.ErrorMessage,
	})
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
	err := r.queries.UpdateReminderEvent(ctx, sqldb.UpdateReminderEventParams{
		SentAt:       pgconv.TimestamptzFromTimePtr(event.SentAt),
		Status:       sqldb.ReminderEventStatus(event.Status),
		ErrorMessage: event.ErrorMessage,
		ID:           event.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to update reminder event: %w", err)
	}
	return nil
}

// GetPendingEvents retrieves pending reminder events.
func (r *reminderRepository) GetPendingEvents(ctx context.Context, before time.Time) ([]model.ReminderEvent, error) {
	events, err := r.queries.GetPendingReminderEvents(ctx, sqldb.GetPendingReminderEventsParams{ScheduledAt: pgconv.TimestamptzFromTime(before)})
	if err != nil {
		return nil, fmt.Errorf("failed to get pending events: %w", err)
	}
	return toModelReminderEvents(events), nil
}

func toModelReminderRules(reminders []sqldb.ReminderRule) []model.ReminderRule {
	models := make([]model.ReminderRule, 0, len(reminders))
	for _, reminder := range reminders {
		models = append(models, toModelReminderRule(reminder))
	}
	return models
}

func toModelReminderRule(reminder sqldb.ReminderRule) model.ReminderRule {
	return model.ReminderRule{
		ID:               reminder.ID,
		DocumentID:       reminder.DocumentID,
		TenantID:         reminder.TenantID,
		RuleType:         reminder.RuleType,
		TriggerDate:      reminder.TriggerDate.Time,
		NotifyDaysBefore: pgconv.IntsFromInt32s(reminder.NotifyDaysBefore),
		Source:           string(reminder.Source),
		Active:           reminder.Active,
		CreatedAt:        reminder.CreatedAt.Time,
	}
}

func toModelReminderEvents(events []sqldb.ReminderEvent) []model.ReminderEvent {
	models := make([]model.ReminderEvent, 0, len(events))
	for _, event := range events {
		models = append(models, model.ReminderEvent{
			ID:           event.ID,
			RuleID:       event.RuleID,
			ScheduledAt:  event.ScheduledAt.Time,
			SentAt:       pgconv.TimePtrFromTimestamptz(event.SentAt),
			Channel:      event.Channel,
			Status:       model.ReminderEventStatus(event.Status),
			ErrorMessage: event.ErrorMessage,
			CreatedAt:    event.CreatedAt.Time,
		})
	}
	return models
}
