package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/labstack/gommon/log"
	"github.com/letenk/golang-authentication/bob/models"
	"github.com/letenk/golang-authentication/exceptions"
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
	// Check if email already exists
	if param.Email != "" {
		existingUser, err := service.userRepository.FindByEmail(ctx, param.Email)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Errorf("failed to check email: %s", err)
			return nil, exceptions.NewBusinessLogicError(exceptions.DataCreateFailed, errors.New("failed to check email"), nil)
		}
		if existingUser != nil {
			return nil, exceptions.NewBusinessLogicError(exceptions.DataConflict, errors.New("email already registered"), nil)
		}
	}

	// Check if phone already exists
	if param.Phone != "" {
		cleanedPhone := utils.EnsurePhoneNumber(param.Phone)
		param.Phone = cleanedPhone

		existingUser, err := service.userRepository.FindByPhone(ctx, param.Phone)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Errorf("failed to check phone: %s", err)
			return nil, exceptions.NewBusinessLogicError(exceptions.DataCreateFailed, errors.New("failed to check phone"), nil)
		}
		if existingUser != nil {
			return nil, exceptions.NewBusinessLogicError(exceptions.DataConflict, errors.New("phone already registered"), nil)
		}
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(param.Password)
	if err != nil {
		log.Errorf("failed to hash password: %s", err)
		return nil, exceptions.NewBusinessLogicError(exceptions.DataCreateFailed, errors.New("failed to hash password"), nil)
	}

	// Prepare user data
	data := models.UserSetter{
		Password:  omit.From(hashedPassword),
		CreatedBy: omit.From(int64(0)),
		UpdatedBy: omit.From(int64(0)),
	}

	// Set optional fields
	if param.FullName != "" {
		data.Name = omitnull.From(param.FullName)
	}

	if param.Phone != "" {
		data.Phone = omitnull.From(param.Phone)
	}

	if param.Email != "" {
		data.Email = omitnull.From(param.Email)
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
		log.Errorf("failed to register user: %s", err)
		return nil, exceptions.NewBusinessLogicError(exceptions.BadData, errors.New("failed to register user"), nil)
	}

	return result, nil
}
