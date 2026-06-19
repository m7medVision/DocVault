package app

import "errors"

var (
	ErrReminderNotFound    = errors.New("reminder not found")
	ErrInvalidReminderDate = errors.New("invalid reminder date")
	ErrReminderNotActive   = errors.New("reminder is not active")
)
