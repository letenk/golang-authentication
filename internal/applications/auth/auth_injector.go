//go:build wireinject
// +build wireinject

package auth

import (
	"github.com/google/wire"

	"github.com/letenk/golang-authentication/configs/credential"
	"github.com/letenk/golang-authentication/configs/database"
	"github.com/letenk/golang-authentication/configs/jwt_config"
	authService "github.com/letenk/golang-authentication/internal/applications/auth/service"
	emailSvc "github.com/letenk/golang-authentication/internal/applications/email/service"
	otpRepo "github.com/letenk/golang-authentication/internal/applications/otp/repository/db"
	tokenRepo "github.com/letenk/golang-authentication/internal/applications/refresh_token/repository/db"
	"github.com/letenk/golang-authentication/internal/applications/transaction"
	userRepo "github.com/letenk/golang-authentication/internal/applications/user/repository/db"
)

var providerSetAuth = wire.NewSet(
	userRepo.NewUserRepository,
	wire.Bind(new(userRepo.UserRepository), new(*userRepo.UserRepositoryImpl)),

	tokenRepo.NewRefreshTokenRepository,
	wire.Bind(new(tokenRepo.RefreshTokenRepository), new(*tokenRepo.RefreshTokenRepositoryImpl)),

	otpRepo.NewOTPRepository,
	wire.Bind(new(otpRepo.OTPRepository), new(*otpRepo.OTPRepositoryImpl)),

	transaction.NewTrxService,

	authService.NewAuthService,
	wire.Bind(new(authService.AuthService), new(*authService.AuthServiceImpl)),
)

// InitializeAuthService initializes all dependencies for AuthController
func InitializeAuthService(
	db *database.BobDB,
	jwtConfig *jwt_config.JWTConfig,
	emailService emailSvc.EmailService,
	otpConfig *credential.OTPConfig,
) *authService.AuthServiceImpl {
	wire.Build(providerSetAuth)
	return nil
}
