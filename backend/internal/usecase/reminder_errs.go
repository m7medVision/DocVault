package usecase

import "errors"

var (
	ErrReminderNotFound    = errors.New("reminder not found")
	ErrInvalidReminderDate = errors.New("invalid reminder date")
	ErrReminderNotActive   = errors.New("reminder is not active")

	ErrNotificationNotFound = errors.New("notification not found")

	ErrAuditEventNotFound = errors.New("audit event not found")
)
