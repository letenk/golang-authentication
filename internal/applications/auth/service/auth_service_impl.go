package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/letenk/golang-authentication/bob/models"
	"github.com/letenk/golang-authentication/internal/applications/auth/dto"
	"github.com/letenk/golang-authentication/internal/applications/user/repository/db"
	"github.com/letenk/golang-authentication/internal/utils"
)

type AuthServiceImpl struct {
	userRepository db.UserRepository
}

func NewAuthService(
	userRepository db.UserRepository,
) *AuthServiceImpl {
	return &AuthServiceImpl{
		userRepository: userRepository,
	}
}

func (service *AuthServiceImpl) Register(ctx context.Context, param *dto.ParameterRegister) (*models.User, error) {
	// Validate input
	if param.Email == "" && param.Phone == "" {
		return nil, errors.New("email or phone is required")
	}

	if param.Password == "" {
		return nil, errors.New("password is required")
	}

	// Check if email already exists
	if param.Email != "" {
		existingUser, err := service.userRepository.FindByEmail(ctx, param.Email)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("failed to check email: %w", err)
		}
		if existingUser != nil {
			return nil, errors.New("email already registered")
		}
	}

	// Check if phone already exists
	if param.Phone != "" {
		existingUser, err := service.userRepository.FindByPhone(ctx, param.Phone)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("failed to check phone: %w", err)
		}
		if existingUser != nil {
			return nil, errors.New("phone already registered")
		}
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(param.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Prepare user data
	data := models.UserSetter{
		Email:    omit.From(param.Email),
		Password: omit.From(hashedPassword),
	}

	// Set optional fields
	if param.FullName != "" {
		data.Name = omitnull.From(param.FullName)
	}

	if param.Phone != "" {
		data.Phone = omitnull.From(param.Phone)
	}

	// Determine login type
	loginType := "email"
	if param.Email == "" && param.Phone != "" {
		loginType = "phone"
	}
	data.LoginType = omit.From(loginType)

	// Set default values
	data.IsVerified = omitnull.From(false)

	// Create new user
	result, err := service.userRepository.Create(ctx, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return result, nil
}
