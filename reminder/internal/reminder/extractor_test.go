package reminder

import (
	"context"
	"testing"
	"time"
)

func TestReminderExtractorCreatesWarrantyRuleFromExtractedExpiry(t *testing.T) {
	expiryDate := time.Now().AddDate(1, 0, 0)
	dates := &ExtractedDates{ExpiryDate: &expiryDate}

	extractor := NewReminderExtractor([]int{30, 7, 1})
	rules, err := extractor.ExtractRules(context.Background(), "doc-1", "tenant-1", dates, "warranty")
	if err != nil {
		t.Fatalf("ExtractRules() unexpected error: %v", err)
	}

	if len(rules) != 1 {
		t.Fatalf("ExtractRules() created %d rules, want 1", len(rules))
	}

	rule := rules[0]
	if rule.RuleType != RuleTypeExpiry {
		t.Fatalf("rule type = %q, want %q", rule.RuleType, RuleTypeExpiry)
	}
	if len(rule.NotifyDaysBefore) == 0 || rule.NotifyDaysBefore[0] != 0 {
		t.Fatalf("notify days = %v, want immediate warranty reminder", rule.NotifyDaysBefore)
	}
}
