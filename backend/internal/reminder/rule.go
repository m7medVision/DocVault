package reminder

import (
	"errors"
	"time"
)

// Reminder construction invariants.
var (
	ErrMissingTenant      = errors.New("tenant ID is required")
	ErrMissingDocument    = errors.New("document ID is required")
	ErrMissingRuleType    = errors.New("rule type is required")
	ErrInvalidTriggerDate = errors.New("invalid reminder date")
	ErrNegativeNotifyDays = errors.New("notify_days_before must be non-negative")
)

// NewReminderRuleParams carries the inputs for constructing a ReminderRule.
type NewReminderRuleParams struct {
	ID               string
	TenantID         string
	DocumentID       string
	RuleType         string
	TriggerDate      time.Time
	NotifyDaysBefore []int
	Source           string
}

// NewReminderRule constructs an active ReminderRule, enforcing the invariants a
// rule must satisfy: a tenant, a document, a non-empty rule type, a real
// trigger date, and non-negative notify offsets.
func NewReminderRule(p NewReminderRuleParams) (*ReminderRule, error) {
	if p.TenantID == "" {
		return nil, ErrMissingTenant
	}
	if p.DocumentID == "" {
		return nil, ErrMissingDocument
	}
	if p.RuleType == "" {
		return nil, ErrMissingRuleType
	}
	if p.TriggerDate.IsZero() {
		return nil, ErrInvalidTriggerDate
	}
	for _, d := range p.NotifyDaysBefore {
		if d < 0 {
			return nil, ErrNegativeNotifyDays
		}
	}

	return &ReminderRule{
		ID:               p.ID,
		DocumentID:       p.DocumentID,
		TenantID:         p.TenantID,
		RuleType:         p.RuleType,
		TriggerDate:      p.TriggerDate,
		NotifyDaysBefore: p.NotifyDaysBefore,
		Source:           p.Source,
		Active:           true,
		CreatedAt:        time.Now(),
	}, nil
}
