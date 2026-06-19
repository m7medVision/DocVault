package reminder

import (
	"errors"
	"testing"
	"time"
)

func TestNewReminderRule(t *testing.T) {
	trigger := time.Now().Add(24 * time.Hour)
	r, err := NewReminderRule(NewReminderRuleParams{
		ID:               "r1",
		TenantID:         "t1",
		DocumentID:       "d1",
		RuleType:         "expiry",
		TriggerDate:      trigger,
		NotifyDaysBefore: []int{30, 7, 1},
		Source:           "manual",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !r.Active {
		t.Error("new rule must be active")
	}
	if r.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
	if !r.TriggerDate.Equal(trigger) {
		t.Error("TriggerDate not preserved")
	}
}

func TestNewReminderRuleInvariants(t *testing.T) {
	base := NewReminderRuleParams{TenantID: "t1", DocumentID: "d1", RuleType: "expiry", TriggerDate: time.Now()}

	cases := []struct {
		name   string
		mutate func(p *NewReminderRuleParams)
		want   error
	}{
		{"missing tenant", func(p *NewReminderRuleParams) { p.TenantID = "" }, ErrMissingTenant},
		{"missing document", func(p *NewReminderRuleParams) { p.DocumentID = "" }, ErrMissingDocument},
		{"missing rule type", func(p *NewReminderRuleParams) { p.RuleType = "" }, ErrMissingRuleType},
		{"zero trigger date", func(p *NewReminderRuleParams) { p.TriggerDate = time.Time{} }, ErrInvalidTriggerDate},
		{"negative notify", func(p *NewReminderRuleParams) { p.NotifyDaysBefore = []int{-1} }, ErrNegativeNotifyDays},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := base
			tc.mutate(&p)
			if _, err := NewReminderRule(p); !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want %v", err, tc.want)
			}
		})
	}
}
