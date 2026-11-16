package database

import (
	"database/sql"
	"time"
)

// Conversions between database/sql nullable types and the plain Go values the
// domain uses. Kept here so every repository maps NULL the same way.

// TimePtr returns nil for NULL, otherwise a pointer to the time.
func TimePtr(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

// NullString stores "" as NULL.
func NullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

// NullStringPtr stores nil as NULL (used for "leave unchanged" in partial updates).
func NullStringPtr(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}
