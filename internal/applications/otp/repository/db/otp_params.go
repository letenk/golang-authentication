package db

import "time"

type CreateOTPParams struct {
	UserID    int64
	Code      string
	Purpose   OTPPurpose
	ExpiresAt time.Time
}
