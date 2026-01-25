package service

import (
	"context"

	"github.com/letenk/golang-authentication/bob/models"
	"github.com/letenk/golang-authentication/internal/applications/auth/dto"
)


type AuthService interface {
	Register(ctx context.Context, param *dto.ParameterRegister)(*models.User, error)
}