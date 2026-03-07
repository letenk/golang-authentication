package repository

import (
	"context"

	"github.com/letenk/golang-authentication/bob/models"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *models.RefreshTokenSetter) (*models.RefreshToken, error)
	FindByToken(ctx context.Context, token string) (*models.RefreshToken, error)
	FindActiveByUserID(ctx context.Context, userID int64) ([]*models.RefreshToken, error)
	CountActiveByUserID(ctx context.Context, userID int64) (int, error)
	Revoke(ctx context.Context, token string) error
	ReplaceToken(ctx context.Context, oldToken, newToken string) error
	RevokeAllByUserID(ctx context.Context, userID int64) (int, error)
}
