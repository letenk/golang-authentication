//go:build wireinject
// +build wireinject

package auth

import (
	"github.com/google/wire"

	"github.com/letenk/golang-authentication/configs/jwt_config"
	"github.com/letenk/golang-authentication/configs/database"
	authService "github.com/letenk/golang-authentication/internal/applications/auth/service"
	tokenRepo "github.com/letenk/golang-authentication/internal/applications/refresh_token/repository/db"
	userRepo "github.com/letenk/golang-authentication/internal/applications/user/repository/db"
)

var providerSetAuth = wire.NewSet(
	userRepo.NewUserRepository,
	wire.Bind(new(userRepo.UserRepository), new(*userRepo.UserRepositoryImpl)),

	tokenRepo.NewRefreshTokenRepository,
	wire.Bind(new(tokenRepo.RefreshTokenRepository), new(*tokenRepo.RefreshTokenRepositoryImpl)),

	authService.NewAuthService,
	wire.Bind(new(authService.AuthService), new(*authService.AuthServiceImpl)),
)

// InitializeAuthService initializes all dependencies for AuthController
func InitializeAuthService(db *database.BobDB, jwtConfig *jwt_config.JWTConfig) *authService.AuthServiceImpl {
	wire.Build(providerSetAuth)
	return nil
}
