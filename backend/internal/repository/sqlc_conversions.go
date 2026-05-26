package repository

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func timestamptzFromTime(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func timestamptzFromTimePtr(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return timestamptzFromTime(*t)
}

func timePtrFromTimestamptz(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func dateFromTime(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}

func intsFromInt32s(values []int32) []int {
	ints := make([]int, 0, len(values))
	for _, value := range values {
		ints = append(ints, int(value))
	}
	return ints
}

func int32sFromInts(values []int) []int32 {
	ints := make([]int32, 0, len(values))
	for _, value := range values {
		ints = append(ints, int32(value))
	}
	return ints
}
