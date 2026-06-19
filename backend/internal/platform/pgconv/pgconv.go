// Package pgconv holds small conversions between Go values and the pgtype
// wrappers sqlc generates. They are shared by the per-context postgres adapters
// so each slice does not have to re-implement the same nil-aware timestamp and
// slice conversions.
package pgconv

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// TimestamptzFromTime wraps a time.Time as a valid timestamptz.
func TimestamptzFromTime(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// TimestamptzFromTimePtr wraps an optional time.Time; a nil pointer yields a
// NULL timestamptz.
func TimestamptzFromTimePtr(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return TimestamptzFromTime(*t)
}

// TimePtrFromTimestamptz unwraps a timestamptz into an optional time.Time; an
// invalid (NULL) value yields nil.
func TimePtrFromTimestamptz(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

// DateFromTime wraps a time.Time as a valid date.
func DateFromTime(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}

// IntsFromInt32s widens a slice of int32 to int.
func IntsFromInt32s(values []int32) []int {
	ints := make([]int, 0, len(values))
	for _, value := range values {
		ints = append(ints, int(value))
	}
	return ints
}

// Int32sFromInts narrows a slice of int to int32.
func Int32sFromInts(values []int) []int32 {
	ints := make([]int32, 0, len(values))
	for _, value := range values {
		ints = append(ints, int32(value))
	}
	return ints
}
