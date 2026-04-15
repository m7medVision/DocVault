// Package reminder provides reminder extraction and management.
package reminder

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"
)

const (
	RuleTypeExpiry  = "expiry"
	RuleTypeRenewal = "renewal"
	RuleTypeDueDate = "due_date"
	RuleTypeOther   = "other"
)

const (
	SourceAuto   = "auto"
	SourceManual = "manual"
)

// ReminderRule represents a reminder rule to be created.
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

// ExtractedDates holds dates extracted from document text.
type ExtractedDates struct {
	IssueDate   *time.Time
	ExpiryDate  *time.Time
	DueDate     *time.Time
	RenewalDate *time.Time
	DatesFound  []DateMatch
}

// DateMatch represents a matched date with context.
type DateMatch struct {
	Date       time.Time
	Pattern    string
	Context    string
	Confidence float64
}

// DateExtractor interface for extracting dates from document text.
type DateExtractor interface {
	ExtractDates(ctx context.Context, text string, docType string) (*ExtractedDates, error)
}

// ReminderExtractor creates reminder rules from extracted dates.
type ReminderExtractor struct {
	notifyDaysBefore []int
}

// NewReminderExtractor creates a new reminder extractor.
func NewReminderExtractor(notifyDaysBefore []int) *ReminderExtractor {
	if len(notifyDaysBefore) == 0 {
		notifyDaysBefore = []int{30, 7, 1}
	}
	return &ReminderExtractor{notifyDaysBefore: notifyDaysBefore}
}

// ExtractRules creates reminder rules from extracted dates.
func (e *ReminderExtractor) ExtractRules(
	ctx context.Context,
	documentID, tenantID string,
	dates *ExtractedDates,
	docType string,
) ([]ReminderRule, error) {
	var rules []ReminderRule

	// Priority mapping by document type
	priorityMap := map[string]bool{
		"contract": true,
		"warranty": true,
		"identity": true,
	}

	// Determine the primary date for reminders
	var primaryDate *time.Time
	var ruleType string

	if dates.ExpiryDate != nil {
		primaryDate = dates.ExpiryDate
		ruleType = RuleTypeExpiry
	} else if dates.RenewalDate != nil {
		primaryDate = dates.RenewalDate
		ruleType = RuleTypeRenewal
	} else if dates.DueDate != nil {
		primaryDate = dates.DueDate
		ruleType = RuleTypeDueDate
	}

	// Create rule if we have a valid future date
	if primaryDate != nil && primaryDate.After(time.Now()) {
		rule := ReminderRule{
			ID:               generateID(),
			DocumentID:       documentID,
			TenantID:         tenantID,
			RuleType:         ruleType,
			TriggerDate:      *primaryDate,
			NotifyDaysBefore: e.notifyDaysBefore,
			Source:           SourceAuto,
			Active:           true,
		}

		// Higher priority documents get immediate reminders
		if priorityMap[docType] {
			rule.NotifyDaysBefore = append([]int{0}, rule.NotifyDaysBefore...)
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

// generateID creates a unique ID for a rule.
func generateID() string {
	return newUUIDString()
}

func newUUIDString() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("00000000-0000-4000-8000-%012x", time.Now().UnixNano()&0xffffffffffff)
	}

	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// GetBestDate returns the most likely reminder date.
func (d *ExtractedDates) GetBestDate() *time.Time {
	if d.ExpiryDate != nil {
		return d.ExpiryDate
	}
	if d.RenewalDate != nil {
		return d.RenewalDate
	}
	if d.DueDate != nil {
		return d.DueDate
	}
	return nil
}
