package app

import "errors"

// ErrAuditEventNotFound is returned when no audit event matches the lookup.
var ErrAuditEventNotFound = errors.New("audit event not found")
