package app

import (
	"context"

	model "github.com/docvault/backend/internal/reminder"
)

// ReminderStore is the persistence port the HTTP-facing ReminderService needs.
// It is the rule-management subset of repository.ReminderRepository; the event
// dispatch methods (CreateEvent/UpdateEvent/GetPendingEvents/ListUpcoming) are
// driven by the separate reminder worker, not this service.
type ReminderStore interface {
	Create(ctx context.Context, rule *model.ReminderRule) error
	GetByID(ctx context.Context, tenantID, id string) (*model.ReminderRule, error)
	GetByDocument(ctx context.Context, tenantID, documentID string) ([]model.ReminderRule, error)
	ListByTenant(ctx context.Context, tenantID string, activeOnly bool) ([]model.ReminderRule, error)
	Update(ctx context.Context, rule *model.ReminderRule) error
}
