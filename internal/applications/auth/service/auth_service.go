package service

import (
	"context"

	"github.com/letenk/golang-authentication/bob/models"
	"github.com/letenk/golang-authentication/internal/applications/auth/dto"
)

type AuthService interface {
	Register(ctx context.Context, param *dto.ParameterRegister) (*models.User, error)
	Login(ctx context.Context, req *dto.LoginRequest, ipAddress, userAgent string) (*dto.LoginResponse, error)
	RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest, ipAddress, userAgent string) (*dto.RefreshTokenResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	LogoutAll(ctx context.Context, userID int64) (int, error)
	GetMe(ctx context.Context, userID int64) (*dto.GetMeResponse, error)
	DeleteAccount(ctx context.Context, userID int64) error
}
