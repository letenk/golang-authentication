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
	GetSessions(ctx context.Context, userID int64) ([]*dto.ActiveSessionResponse, error)
	RevokeSession(ctx context.Context, userID int64, sessionID int64) error
	UpdateProfile(ctx context.Context, userID int64, req *dto.UpdateProfileRequest) (*dto.UserResponse, error)
	UpdatePassword(ctx context.Context, userID int64, req *dto.UpdatePasswordRequest) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error
}
