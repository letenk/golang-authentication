package repository

import (
	"context"

	"github.com/letenk/golang-authentication/internal/domain"
)

// RefreshTokenRepository stores login sessions. "Active" means not revoked and
// not expired. Methods return domain.ErrNotFound when no row matches or an
// update affects no row. When ctx comes from TrxService.WithTx, methods run
// inside that transaction.
type RefreshTokenRepository interface {
	Create(ctx context.Context, params CreateRefreshTokenParams) (*domain.RefreshToken, error)
	FindByToken(ctx context.Context, token string) (*domain.RefreshToken, error)
	// FindActiveByUserID returns active sessions, newest first.
	FindActiveByUserID(ctx context.Context, userID int64) ([]*domain.RefreshToken, error)
	CountActiveByUserID(ctx context.Context, userID int64) (int, error)
	Revoke(ctx context.Context, token string) error
	// ReplaceToken revokes oldToken and records newToken as its replacement,
	// only if oldToken is not revoked yet; otherwise it returns ErrNotFound.
	ReplaceToken(ctx context.Context, oldToken, newToken string) error
	// RevokeAllByUserID revokes every unrevoked session and returns how many.
	RevokeAllByUserID(ctx context.Context, userID int64) (int, error)
	// RevokeByIDForUser revokes one unrevoked session owned by userID.
	RevokeByIDForUser(ctx context.Context, sessionID int64, userID int64) error
}
