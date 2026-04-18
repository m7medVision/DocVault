// Package model defines the domain models for DocVault.
package model

import (
	"time"
)

// Tenant represents a root isolation unit.
type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Plan      string    `json:"plan"` // free | personal | business
	CreatedAt time.Time `json:"created_at"`
}

// Organization represents a sub-unit within a tenant.
type Organization struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// User represents a user account with internal authentication.
type User struct {
	ID                  string     `json:"id"`
	TenantID            string     `json:"tenant_id"`
	Email               string     `json:"email"`
	PasswordHash        string     `json:"-"` // Never expose in JSON
	DisplayName         string     `json:"display_name"`
	Locale              string     `json:"locale"` // ar | en
	EmailVerified       bool       `json:"email_verified"`
	LastLoginAt         *time.Time `json:"last_login_at,omitempty"`
	FailedLoginAttempts int        `json:"-"` // Never expose in JSON
	LockedUntil         *time.Time `json:"-"` // Never expose in JSON
	CreatedAt           time.Time  `json:"created_at"`
}

// Membership links a user to an organization with a role.
type Membership struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	OrgID     string    `json:"org_id"`
	Role      string    `json:"role"` // owner | admin | member | viewer
	CreatedAt time.Time `json:"created_at"`
}

// DocumentStatus represents the coarse processing state of a document.
type DocumentStatus string

const (
	DocumentStatusPending    DocumentStatus = "pending"
	DocumentStatusProcessing DocumentStatus = "processing"
	DocumentStatusProcessed  DocumentStatus = "processed"
	DocumentStatusFailed     DocumentStatus = "failed"
)

// ProcessingStage represents the granular pipeline stage of a document.
type ProcessingStage string

const (
	StageUploaded          ProcessingStage = "uploaded"
	StageOCRQueued         ProcessingStage = "ocr_queued"
	StageOCRRunning        ProcessingStage = "ocr_running"
	StageOCRComplete       ProcessingStage = "ocr_complete"
	StageProcessingQueued  ProcessingStage = "processing_queued"
	StageProcessingRunning ProcessingStage = "processing_running"
	StageIndexing          ProcessingStage = "indexing"
	StageSuggesting        ProcessingStage = "suggesting"
	StageCompleted         ProcessingStage = "completed"
	StageOCRFailed         ProcessingStage = "ocr_failed"
	StageProcessingFailed  ProcessingStage = "processing_failed"
)

// Document represents a stable identity record for an uploaded file.
type Document struct {
	ID              string         `json:"id"`
	TenantID        string         `json:"tenant_id"`
	OrgID           string         `json:"org_id"`
	FolderID        *string        `json:"folder_id,omitempty"`
	OwnerID         string         `json:"owner_id"`
	Title           string         `json:"title"`
	DocType         string         `json:"doc_type"`
	Status          DocumentStatus `json:"status"`
	Language        *string        `json:"language,omitempty"`
	ProcessingStage *string        `json:"processing_stage,omitempty"`
	ProcessingError *string        `json:"processing_error,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
}

// DocumentVersion represents a single version of a document.
type DocumentVersion struct {
	ID            string    `json:"id"`
	DocumentID    string    `json:"document_id"`
	VersionNumber int       `json:"version_number"`
	StorageKey    string    `json:"storage_key"`
	MimeType      string    `json:"mime_type"`
	FileSize      int64     `json:"file_size"`
	UploadedBy    *string   `json:"uploaded_by,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// DocumentPage represents OCR output for a single page.
type DocumentPage struct {
	ID             string    `json:"id"`
	DocumentID     string    `json:"document_id"`
	VersionID      string    `json:"version_id"`
	PageNumber     int       `json:"page_number"`
	OCRText        *string   `json:"ocr_text,omitempty"`
	TranslatedText *string   `json:"translated_text,omitempty"`
	Confidence     *float32  `json:"confidence,omitempty"`
	OCRModel       string    `json:"ocr_model"`
	CreatedAt      time.Time `json:"created_at"`
}

// DocumentMetadata represents key-value metadata for a document.
type DocumentMetadata struct {
	ID             string     `json:"id"`
	DocumentID     string     `json:"document_id"`
	Key            string     `json:"key"`
	ExtractedValue *string    `json:"extracted_value,omitempty"`
	CorrectedValue *string    `json:"corrected_value,omitempty"`
	CorrectedBy    *string    `json:"corrected_by,omitempty"`
	CorrectedAt    *time.Time `json:"corrected_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// ExtractedTextChunk represents a retrieval unit with embedding.
type ExtractedTextChunk struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"document_id"`
	PageID     string    `json:"page_id"`
	ChunkIndex int       `json:"chunk_index"`
	ChunkText  string    `json:"chunk_text"`
	CreatedAt  time.Time `json:"created_at"`
}

// Folder represents a nested folder tree node.
type Folder struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	OrgID     string    `json:"org_id"`
	ParentID  *string   `json:"parent_id,omitempty"`
	Name      string    `json:"name"`
	CreatedBy *string   `json:"created_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Tag represents a free-form label.
type Tag struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// ReminderRule defines when to send reminders for a document.
type ReminderRule struct {
	ID               string    `json:"id"`
	DocumentID       string    `json:"document_id"`
	TenantID         string    `json:"tenant_id"`
	RuleType         string    `json:"rule_type"` // expiry, renewal, due_date
	TriggerDate      time.Time `json:"trigger_date"`
	NotifyDaysBefore []int     `json:"notify_days_before"`
	Source           string    `json:"source"` // auto | manual
	Active           bool      `json:"active"`
	CreatedAt        time.Time `json:"created_at"`
}

// ReminderEvent tracks notification delivery state.
type ReminderEventStatus string

const (
	ReminderEventStatusPending ReminderEventStatus = "pending"
	ReminderEventStatusSent    ReminderEventStatus = "sent"
	ReminderEventStatusFailed  ReminderEventStatus = "failed"
	ReminderEventStatusSnoozed ReminderEventStatus = "snoozed"
)

type ReminderEvent struct {
	ID           string              `json:"id"`
	RuleID       string              `json:"rule_id"`
	ScheduledAt  time.Time           `json:"scheduled_at"`
	SentAt       *time.Time          `json:"sent_at,omitempty"`
	Channel      string              `json:"channel"` // in_app | email
	Status       ReminderEventStatus `json:"status"`
	ErrorMessage *string             `json:"error_message,omitempty"`
	CreatedAt    time.Time           `json:"created_at"`
}

// AuditEvent records an action in the system (append-only).
type AuditEvent struct {
	ID         string                 `json:"id"`
	TenantID   string                 `json:"tenant_id"`
	ActorID    *string                `json:"actor_id,omitempty"`
	EntityType string                 `json:"entity_type"`
	EntityID   string                 `json:"entity_id"`
	Action     string                 `json:"action"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

// NotificationType represents the type of notification.
type NotificationType string

const (
	NotificationTypeReminder NotificationType = "reminder"
	NotificationTypeSystem   NotificationType = "system"
)

// NotificationStatus represents the read state of a notification.
type NotificationStatus string

const (
	NotificationStatusUnread NotificationStatus = "unread"
	NotificationStatusRead   NotificationStatus = "read"
)

// Notification represents an in-app notification for a user.
type Notification struct {
	ID        string                 `json:"id"`
	TenantID  string                 `json:"tenant_id"`
	UserID    string                 `json:"user_id"`
	Type      NotificationType       `json:"type"`
	Title     string                 `json:"title"`
	Body      string                 `json:"body"`
	Link      *string                `json:"link,omitempty"`
	Status    NotificationStatus     `json:"status"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	ReadAt    *time.Time             `json:"read_at,omitempty"`
}
