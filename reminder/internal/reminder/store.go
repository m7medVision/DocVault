package reminder

import (
	"context"
	"time"
)

type ReminderRuleRepository interface {
	Create(ctx context.Context, rule *ReminderRule) error
	Update(ctx context.Context, rule *ReminderRule) error
	FindByUser(ctx context.Context, userID string) ([]*ReminderRule, error)
	FindUpcoming(ctx context.Context, userID string, before time.Time) ([]*ReminderRule, error)
	Delete(ctx context.Context, id string) error
}

type ReminderEventRepository interface {
	Create(ctx context.Context, event *ReminderEvent) error
	MarkSent(ctx context.Context, id string) error
	FindPending(ctx context.Context, before time.Time) ([]*ReminderEvent, error)
}

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }

type EmailSender interface {
	Send(ctx context.Context, to, subject, body string) error
}

type ReminderDeriver interface {
	DeriveRules(extractedText string, language string, docID string) ([]*ReminderRule, error)
}
