package service

import (
	"context"

	"github.com/letenk/golang-authentication/bob/models"
	"github.com/letenk/golang-authentication/internal/applications/auth/dto"
)

type AuthService interface {
	Register(ctx context.Context, param *dto.ParameterRegister) (*models.User, error)
	Login(ctx context.Context, req *dto.LoginRequest, ipAddress, userAgent string) (*dto.LoginResponse, error)
	// RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, error)
	// Logout(ctx context.Context, req *dto.LogoutRequest, userID int64) error
	// LogoutAll(ctx context.Context, userID int64) (*dto.LogoutAllResponse, error)
	// GetMe(ctx context.Context, userID int64) (*dto.GetMeResponse, error)
}
