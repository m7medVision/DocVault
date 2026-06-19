package app

import "errors"

// ErrNotificationNotFound is returned when no notification matches the lookup.
var ErrNotificationNotFound = errors.New("notification not found")
