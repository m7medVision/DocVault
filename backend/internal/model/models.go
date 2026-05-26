// Package model preserves the old import path while domain types move into
// bounded-context packages under internal/domain.
package model

import (
	"github.com/docvault/backend/internal/domain/audit"
	"github.com/docvault/backend/internal/domain/document"
	"github.com/docvault/backend/internal/domain/identity"
	"github.com/docvault/backend/internal/domain/notification"
	"github.com/docvault/backend/internal/domain/reminder"
)

type Tenant = identity.Tenant
type Organization = identity.Organization
type User = identity.User
type Membership = identity.Membership

type DocumentStatus = document.DocumentStatus

const (
	DocumentStatusPending    = document.DocumentStatusPending
	DocumentStatusProcessing = document.DocumentStatusProcessing
	DocumentStatusProcessed  = document.DocumentStatusProcessed
	DocumentStatusFailed     = document.DocumentStatusFailed
)

type ProcessingStage = document.ProcessingStage

const (
	StageUploaded          = document.StageUploaded
	StageOCRQueued         = document.StageOCRQueued
	StageOCRRunning        = document.StageOCRRunning
	StageOCRComplete       = document.StageOCRComplete
	StageProcessingQueued  = document.StageProcessingQueued
	StageProcessingRunning = document.StageProcessingRunning
	StageIndexing          = document.StageIndexing
	StageSuggesting        = document.StageSuggesting
	StageCompleted         = document.StageCompleted
	StageOCRFailed         = document.StageOCRFailed
	StageProcessingFailed  = document.StageProcessingFailed
)

type Document = document.Document
type DocumentVersion = document.DocumentVersion
type DocumentPage = document.DocumentPage
type DocumentMetadata = document.DocumentMetadata
type ExtractedTextChunk = document.ExtractedTextChunk
type Folder = document.Folder
type Tag = document.Tag
type DocumentStats = document.DocumentStats

type ReminderRule = reminder.ReminderRule
type ReminderEventStatus = reminder.ReminderEventStatus

const (
	ReminderEventStatusPending = reminder.ReminderEventStatusPending
	ReminderEventStatusSent    = reminder.ReminderEventStatusSent
	ReminderEventStatusFailed  = reminder.ReminderEventStatusFailed
	ReminderEventStatusSnoozed = reminder.ReminderEventStatusSnoozed
)

type ReminderEvent = reminder.ReminderEvent
type AuditEvent = audit.AuditEvent

type NotificationType = notification.NotificationType

const (
	NotificationTypeReminder = notification.NotificationTypeReminder
	NotificationTypeSystem   = notification.NotificationTypeSystem
)

type NotificationStatus = notification.NotificationStatus

const (
	NotificationStatusUnread = notification.NotificationStatusUnread
	NotificationStatusRead   = notification.NotificationStatusRead
)

type Notification = notification.Notification
