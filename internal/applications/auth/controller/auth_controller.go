package controller

import (
	"github.com/labstack/echo/v4"
	"github.com/letenk/golang-authentication/internal/applications/auth/dto"
	"github.com/letenk/golang-authentication/internal/applications/auth/service"
	"github.com/letenk/golang-authentication/internal/helper"
	"github.com/letenk/golang-authentication/internal/helper/response"
	"github.com/letenk/golang-authentication/internal/utils"
	"github.com/letenk/golang-authentication/internal/utils/headers"
)

type AuthController struct {
	authService service.AuthService
}

func NewAuthController(
	authService service.AuthService,
) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

// Implement swagger
func (controller *AuthController) Register(ctx echo.Context) error {
	request := &dto.RegisterRequest{}

	err := helper.BindAndValidate(ctx, request)
	if err != nil {
		return err
	}

	// Custom validation: at least email or phone
	if validationErr := request.Validate(); validationErr != nil {
		return validationErr
	}

	param := &dto.ParameterRegister{
		FullName: request.FullName,
		Email:    request.Email,
		Password: request.Password,
		Phone:    request.Phone,
	}

	result, err := controller.authService.Register(ctx.Request().Context(), param)
	if err != nil {
		return err
	}

	successResponse := &dto.RegisterResponse{
		User: &dto.UserResponse{
			ID:         result.ID,
			FullName:   result.Name.GetOr(""),
			Email:      result.Email.GetOr(""),
			Phone:      result.Phone.GetOr(""),
			IsVerified: result.IsVerified.GetOr(false),
		},
	}

	return response.SuccessWithMessage(ctx, "User registered successfully", successResponse)
}

func (controller *AuthController) Login(ctx echo.Context) error {
	request := &dto.LoginRequest{}

	err := helper.BindAndValidate(ctx, request)
	if err != nil {
		return err
	}

	// Get client IP address
	ipAddress := ctx.RealIP()

	// Get user agent
	userAgent := ctx.Request().UserAgent()

	result, err := controller.authService.Login(ctx.Request().Context(), request, ipAddress, userAgent)
	if err != nil {
		return err
	}

	return response.SuccessWithMessage(ctx, "Login successful", result)
}

func (controller *AuthController) GetMe(ctx echo.Context) error {
	// Extract header from context using GetContextHeaders pattern
	header, err := utils.GetContextHeaders(ctx, headers.ContextHeaders)
	if err != nil {
		return err
	}

	result, err := controller.authService.GetMe(ctx.Request().Context(), header.UserID)
	if err != nil {
		return err
	}

	return response.SuccessWithMessage(ctx, "User retrieved successfully", result)
}
