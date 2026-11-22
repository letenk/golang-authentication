package db

import (
	"database/sql"
	"time"

	"github.com/letenk/golang-authentication/internal/domain"
	"github.com/uptrace/bun"
)

// userRow is the persistence shape of the users table. It never leaves this
// package; callers get *domain.User.
type userRow struct {
	bun.BaseModel `bun:"table:users"`

	ID int64 `bun:"id,pk,autoincrement"`
	// nullzero: a zero value is written as DEFAULT (0 / CURRENT_TIMESTAMP).
	Version    int64          `bun:"version,nullzero"`
	CreatedAt  time.Time      `bun:"created_at,nullzero"`
	CreatedBy  int64          `bun:"created_by"`
	UpdatedAt  time.Time      `bun:"updated_at,nullzero"`
	UpdatedBy  int64          `bun:"updated_by"`
	DeletedAt  *time.Time     `bun:"deleted_at"`
	DeletedBy  sql.NullInt64  `bun:"deleted_by"`
	Name       sql.NullString `bun:"name"`
	Email      sql.NullString `bun:"email"`
	Phone      sql.NullString `bun:"phone"`
	Password   string         `bun:"password"`
	LoginType  string         `bun:"login_type"`
	IsVerified sql.NullBool   `bun:"is_verified"`
	VerifiedAt *time.Time     `bun:"verified_at"`
}

func (u *userRow) toDomain() *domain.User {
	return &domain.User{
		ID:         u.ID,
		Name:       u.Name.String,
		Email:      u.Email.String,
		Phone:      u.Phone.String,
		Password:   u.Password,
		LoginType:  u.LoginType,
		IsVerified: u.IsVerified.Bool,
		VerifiedAt: u.VerifiedAt,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}
}
