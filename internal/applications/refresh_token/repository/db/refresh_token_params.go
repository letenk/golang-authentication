package repository

import "time"

// CreateRefreshTokenParams holds the fields for a new session. Device fields
// are stored as given.
type CreateRefreshTokenParams struct {
	UserID     int64
	Token      string
	DeviceName string
	DeviceID   string
	IPAddress  string
	UserAgent  string
	ExpiresAt  time.Time
}
