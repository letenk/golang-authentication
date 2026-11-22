package domain

import "time"

// OTP is a one-time code; Purpose distinguishes its use case
// (e.g. "email_verification", "password_reset").
type OTP struct {
	ID        int64
	UserID    int64
	Code      string
	Purpose   string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}
