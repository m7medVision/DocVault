// Package reminder provides reminder extraction and management.
package reminder

import (
	"context"
	"crypto/rand"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
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
	ExtractDates(text string, language string) ([]ParsedDate, error)
}

// dateExtractor implements DateExtractor for bilingual Arabic + English date extraction.
type dateExtractor struct {
	patterns     []*regexp.Regexp
	arabicMonths map[string]int
}

// ParsedDate represents a date parsed from document text.
type ParsedDate struct {
	Date       time.Time
	Pattern    string
	Context    string
	Confidence float64
}

// NewDateExtractor creates a new date extractor.
func NewDateExtractor() DateExtractor {
	patterns := []*regexp.Regexp{
		// English expiry dates
		regexp.MustCompile(`(?i)(?:expires?|valid\s+until|expir(?:y|ation)\s*date)[:.\s]+(\d{1,2}[/-]\d{1,2}[/-]\d{2,4})`),
		regexp.MustCompile(`(?i)(?:expires?|valid\s+until|expir(?:y|ation)\s*date)[:.\s]+(\d{4}[/-]\d{2}[/-]\d{2})`),
		regexp.MustCompile(`(?i)(?:expires?|valid\s+until)[:.\s]+([A-Za-z]+\s+\d{1,2},?\s+\d{4})`),

		// Issue dates
		regexp.MustCompile(`(?i)(?:dated?|issued?\s*(?:on)?|date\s*of\s*(?:issue|contract))[:.\s]+(\d{1,2}[/-]\d{1,2}[/-]\d{2,4})`),
		regexp.MustCompile(`(?i)(?:dated?|issued?\s*(?:on)?)[:.\s]+(\d{4}[/-]\d{2}[/-]\d{2})`),
		regexp.MustCompile(`(?i)(?:dated?|issued?\s*(?:on)?)[:.\s]+([A-Za-z]+\s+\d{1,2},?\s+\d{4})`),

		// Due dates
		regexp.MustCompile(`(?i)(?:due\s*date|payment\s*due|pay\s*by)[:.\s]+(\d{1,2}[/-]\d{1,2}[/-]\d{2,4})`),
		regexp.MustCompile(`(?i)(?:due\s*date|payment\s*due)[:.\s]+(\d{4}[/-]\d{2}[/-]\d{2})`),

		// Renewal dates
		regexp.MustCompile(`(?i)(?:renew(?:al|s)\s*(?:on|by)?|next\s+renewal)[:.\s]+(\d{1,2}[/-]\d{1,2}[/-]\d{2,4})`),

		// ISO format dates (anywhere in text)
		regexp.MustCompile(`(\d{4}-\d{2}-\d{2})`),

		// Arabic date patterns
		regexp.MustCompile(`(?:تاريخ\s+)?(?:انتهاء|صالح\s+حتى)[:.\s]+(\d{1,2}[/-]\d{1,2}[/-]\d{4})`),
		regexp.MustCompile(`(?:تاريخ\s+)?(?:الإصدار|الإصدار)[:.\s]+(\d{1,2}[/-]\d{1,2}[/-]\d{4})`),
		regexp.MustCompile(`(?:تاريخ\s+)?(?:استحقاق|دفع)[:.\s]+(\d{1,2}[/-]\d{1,2}[/-]\d{4})`),
	}

	arabicMonths := map[string]int{
		"يناير":        1,
		"فبراير":       2,
		"مارس":         3,
		"أبريل":        4,
		"مايو":         5,
		"يونيو":        6,
		"يوليو":        7,
		"أغسطس":        8,
		"سبتمبر":       9,
		"أكتوبر":       10,
		"نوفمبر":       11,
		"ديسمبر":       12,
		"محرم":         1,
		"صفر":          2,
		"ربيع الأول":   3,
		"ربيع الثاني":  4,
		"جمادى الأولى": 5,
		"جمادى الآخرة": 6,
		"رجب":          7,
		"شعبان":        8,
		"رمضان":        9,
		"شوال":         10,
		"ذو القعدة":    11,
		"ذو الحجة":     12,
	}

	return &dateExtractor{
		patterns:     patterns,
		arabicMonths: arabicMonths,
	}
}

// ExtractDates finds all dates in the text (bilingual Arabic + English).
func (e *dateExtractor) ExtractDates(text string, language string) ([]ParsedDate, error) {
	var results []ParsedDate

	isArabic := e.containsArabic(text)

	for _, pattern := range e.patterns {
		matches := pattern.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}

			dateStr := strings.TrimSpace(match[1])

			var parsedDate time.Time
			var err error

			formats := []string{
				"01/02/2006",
				"02/01/2006",
				"01-02-2006",
				"02-01-2006",
				"2006/01/02",
				"2006-01-02",
				"January 2, 2006",
				"Jan 2, 2006",
				"2 January 2006",
				"2 Jan 2006",
			}

			for _, format := range formats {
				parsedDate, err = time.Parse(format, dateStr)
				if err == nil {
					break
				}
			}

			if err != nil && isArabic {
				parsedDate, err = e.parseArabicDate(dateStr)
			}

			if err != nil {
				continue
			}

			if parsedDate.After(time.Now().Add(-24 * time.Hour)) {
				idx := pattern.FindStringIndex(text)
				start := max(0, idx[0]-30)
				end := min(len(text), idx[1]+30)
				context := strings.TrimSpace(text[start:end])

				results = append(results, ParsedDate{
					Date:       parsedDate,
					Pattern:    pattern.String(),
					Context:    context,
					Confidence: 0.8,
				})
			}
		}
	}

	return results, nil
}

// containsArabic checks if text contains Arabic characters.
func (e *dateExtractor) containsArabic(text string) bool {
	for _, r := range text {
		if unicode.Is(unicode.Arabic, r) {
			return true
		}
	}
	return false
}

// parseArabicDate attempts to parse an Arabic date string.
func (e *dateExtractor) parseArabicDate(dateStr string) (time.Time, error) {
	// Try format: day/month/year with Arabic numerals
	arabicNumerals := map[rune]rune{
		'٠': '0', '١': '1', '٢': '2', '٣': '3', '٤': '4',
		'٥': '5', '٦': '6', '٧': '7', '٨': '8', '٩': '9',
	}

	// Convert Arabic numerals to Western
	western := dateStr
	for ar, we := range arabicNumerals {
		western = strings.ReplaceAll(western, string(ar), string(we))
	}

	// Try common formats
	formats := []string{
		"02/01/2006",
		"01/02/2006",
		"2006/01/02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, western); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("failed to parse Arabic date: %s", dateStr)
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
