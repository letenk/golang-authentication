package domain

import "time"

// RefreshToken is one login session on one device.
type RefreshToken struct {
	ID              int64
	UserID          int64
	Token           string
	DeviceName      string
	DeviceID        string
	IPAddress       string
	UserAgent       string
	ExpiresAt       time.Time
	CreatedAt       time.Time
	RevokedAt       *time.Time
	ReplacedByToken string
}

// IsRevoked reports whether the token was revoked (logout or rotation).
func (t *RefreshToken) IsRevoked() bool {
	return t.RevokedAt != nil
}

// IsExpired reports whether the token expired before now.
func (t *RefreshToken) IsExpired(now time.Time) bool {
	return t.ExpiresAt.Before(now)
}
