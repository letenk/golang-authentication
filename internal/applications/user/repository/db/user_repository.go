package db

import (
	"context"
	"time"

	"github.com/letenk/golang-authentication/internal/domain"
)

// UserRepository reads and writes users. Soft-deleted users are invisible to
// every method. Methods return domain.ErrNotFound when no live user matches,
// and domain.ErrEmailTaken / domain.ErrPhoneTaken on unique violations. When
// ctx comes from TrxService.WithTx, methods run inside that transaction.
type UserRepository interface {
	Create(ctx context.Context, params CreateUserParams) (*domain.User, error)
	FindByID(ctx context.Context, id int64) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByPhone(ctx context.Context, phone string) (*domain.User, error)
	UpdateProfile(ctx context.Context, id int64, params UpdateProfileParams) (*domain.User, error)
	UpdatePassword(ctx context.Context, userID int64, hashedPassword string) error
	MarkEmailVerified(ctx context.Context, userID int64, verifiedAt time.Time) error
	SoftDeleteByID(ctx context.Context, id int64, deletedBy int64) error
}
