package repository

import (
	"database/sql"
	"time"

	"github.com/letenk/golang-authentication/internal/domain"
	"github.com/uptrace/bun"
)

// refreshTokenRow is the persistence shape of the refresh_tokens table. It
// never leaves this package; callers get *domain.RefreshToken.
type refreshTokenRow struct {
	bun.BaseModel `bun:"table:refresh_tokens"`

	ID     int64  `bun:"id,pk,autoincrement"`
	UserID int64  `bun:"user_id"`
	Token  string `bun:"token"`
	// Device columns are nullable, but values are stored as given ("" stays
	// ''), so they are written with Valid set; see storedAsGiven.
	DeviceName      sql.NullString `bun:"device_name"`
	DeviceID        sql.NullString `bun:"device_id"`
	IPAddress       sql.NullString `bun:"ip_address"`
	UserAgent       sql.NullString `bun:"user_agent"`
	ExpiresAt       time.Time      `bun:"expires_at"`
	CreatedAt       time.Time      `bun:"created_at,nullzero"` // zero → DEFAULT
	RevokedAt       *time.Time     `bun:"revoked_at"`
	ReplacedByToken sql.NullString `bun:"replaced_by_token"`
}

// storedAsGiven keeps "" as an empty string (not NULL), matching how device
// fields have always been written.
func storedAsGiven(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}

func (t *refreshTokenRow) toDomain() *domain.RefreshToken {
	return &domain.RefreshToken{
		ID:              t.ID,
		UserID:          t.UserID,
		Token:           t.Token,
		DeviceName:      t.DeviceName.String,
		DeviceID:        t.DeviceID.String,
		IPAddress:       t.IPAddress.String,
		UserAgent:       t.UserAgent.String,
		ExpiresAt:       t.ExpiresAt,
		CreatedAt:       t.CreatedAt,
		RevokedAt:       t.RevokedAt,
		ReplacedByToken: t.ReplacedByToken.String,
	}
}
