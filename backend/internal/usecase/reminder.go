// Package usecase contains application business logic.
package usecase

import (
	"context"
	"errors"
	"time"

	model "github.com/docvault/backend/internal/domain/reminder"
	"github.com/docvault/backend/internal/repository"
	"github.com/google/uuid"
)

// ReminderService handles reminder operations.
type ReminderService struct {
	repo repository.ReminderRepository
}

// NewReminderService creates a new ReminderService.
func NewReminderService(repo repository.ReminderRepository) *ReminderService {
	return &ReminderService{repo: repo}
}

// ListRemindersInput contains filters for listing reminders.
type ListRemindersInput struct {
	TenantID   string
	DocumentID string
	ActiveOnly bool
}

// ListRemindersOutput contains the result of listing reminders.
type ListRemindersOutput struct {
	Reminders []model.ReminderRule
}

var errTenantIDRequired = errors.New("tenant ID is required")

// List retrieves reminders for a tenant.
func (s *ReminderService) List(ctx context.Context, input *ListRemindersInput) (*ListRemindersOutput, error) {
	if input.TenantID == "" {
		return nil, errTenantIDRequired
	}

	if s.repo == nil {
		return &ListRemindersOutput{Reminders: []model.ReminderRule{}}, nil
	}

	if input.DocumentID != "" {
		reminders, err := s.repo.GetByDocument(ctx, input.TenantID, input.DocumentID)
		if err != nil {
			return nil, err
		}
		if input.ActiveOnly {
			filtered := make([]model.ReminderRule, 0, len(reminders))
			for _, reminder := range reminders {
				if reminder.Active {
					filtered = append(filtered, reminder)
				}
			}
			reminders = filtered
		}
		return &ListRemindersOutput{Reminders: reminders}, nil
	}

	reminders, err := s.repo.ListByTenant(ctx, input.TenantID, input.ActiveOnly)
	if err != nil {
		return nil, err
	}

	return &ListRemindersOutput{Reminders: reminders}, nil
}

// CreateReminderInput contains the data needed to create a reminder.
type CreateReminderInput struct {
	TenantID         string
	DocumentID       string
	RuleType         string
	TriggerDate      time.Time
	NotifyDaysBefore []int
	Source           string
}

// CreateReminderOutput contains the result of creating a reminder.
type CreateReminderOutput struct {
	Reminder model.ReminderRule
}

// Create creates a new reminder rule.
func (s *ReminderService) Create(ctx context.Context, input *CreateReminderInput) (*CreateReminderOutput, error) {
	reminder, err := model.NewReminderRule(model.NewReminderRuleParams{
		ID:               uuid.New().String(),
		TenantID:         input.TenantID,
		DocumentID:       input.DocumentID,
		RuleType:         input.RuleType,
		TriggerDate:      input.TriggerDate,
		NotifyDaysBefore: input.NotifyDaysBefore,
		Source:           input.Source,
	})
	if err != nil {
		return nil, err
	}

	if s.repo == nil {
		return &CreateReminderOutput{Reminder: *reminder}, nil
	}

	if err := s.repo.Create(ctx, reminder); err != nil {
		return nil, err
	}

	return &CreateReminderOutput{Reminder: *reminder}, nil
}

// UpdateReminderInput contains the data needed to update a reminder.
type UpdateReminderInput struct {
	TenantID string
	ID       string
	Active   *bool
}

// UpdateReminderOutput contains the result of updating a reminder.
type UpdateReminderOutput struct {
	Reminder model.ReminderRule
}

var errReminderIDRequired = errors.New("reminder ID is required")

// Update updates a reminder rule (snooze or dismiss).
func (s *ReminderService) Update(ctx context.Context, input *UpdateReminderInput) (*UpdateReminderOutput, error) {
	if input.TenantID == "" {
		return nil, errTenantIDRequired
	}
	if input.ID == "" {
		return nil, errReminderIDRequired
	}

	if s.repo == nil {
		return nil, errors.New("reminder service not fully implemented")
	}

	reminder, err := s.repo.GetByID(ctx, input.TenantID, input.ID)
	if err != nil {
		if errors.Is(err, repository.ErrReminderNotFound) {
			return nil, ErrReminderNotFound
		}
		return nil, err
	}

	if input.Active != nil {
		reminder.Active = *input.Active
	}

	if err := s.repo.Update(ctx, reminder); err != nil {
		return nil, err
	}

	return &UpdateReminderOutput{Reminder: *reminder}, nil
}
