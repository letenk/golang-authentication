package db

import (
	"time"

	"github.com/letenk/golang-authentication/internal/domain"
	"github.com/uptrace/bun"
)

// otpRow is the persistence shape of the otps table. It never leaves this
// package; callers get *domain.OTP.
type otpRow struct {
	bun.BaseModel `bun:"table:otps"`

	ID        int64      `bun:"id,pk,autoincrement"`
	UserID    int64      `bun:"user_id"`
	Code      string     `bun:"code"`
	Purpose   string     `bun:"purpose"`
	ExpiresAt time.Time  `bun:"expires_at"`
	UsedAt    *time.Time `bun:"used_at"`
	// nullzero: a zero value is written as DEFAULT (CURRENT_TIMESTAMP).
	CreatedAt time.Time `bun:"created_at,nullzero"`
}

func (o *otpRow) toDomain() *domain.OTP {
	return &domain.OTP{
		ID:        o.ID,
		UserID:    o.UserID,
		Code:      o.Code,
		Purpose:   o.Purpose,
		ExpiresAt: o.ExpiresAt,
		UsedAt:    o.UsedAt,
		CreatedAt: o.CreatedAt, // zero when NULL
	}
}
