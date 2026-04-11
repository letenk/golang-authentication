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

	if validationErr := request.Validate(); validationErr != nil {
		return validationErr
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

func (controller *AuthController) RefreshToken(ctx echo.Context) error {
	request := &dto.RefreshTokenRequest{}

	err := helper.BindAndValidate(ctx, request)
	if err != nil {
		return err
	}

	ipAddress := ctx.RealIP()
	userAgent := ctx.Request().UserAgent()

	result, err := controller.authService.RefreshToken(ctx.Request().Context(), request, ipAddress, userAgent)
	if err != nil {
		return err
	}

	return response.SuccessWithMessage(ctx, "Token refreshed successfully", result)
}

func (controller *AuthController) Logout(ctx echo.Context) error {
	request := &dto.LogoutRequest{}

	err := helper.BindAndValidate(ctx, request)
	if err != nil {
		return err
	}

	err = controller.authService.Logout(ctx.Request().Context(), request.RefreshToken)
	if err != nil {
		return err
	}

	return response.SuccessWithMessage(ctx, "Logged out successfully", &dto.LogoutResponse{
		Message: "Logged out successfully",
	})
}

func (controller *AuthController) LogoutAll(ctx echo.Context) error {
	header, err := utils.GetContextHeaders(ctx, headers.ContextHeaders)
	if err != nil {
		return err
	}

	count, err := controller.authService.LogoutAll(ctx.Request().Context(), header.UserID)
	if err != nil {
		return err
	}

	return response.SuccessWithMessage(ctx, "Logged out from all devices successfully", &dto.LogoutAllResponse{
		Message:             "Logged out from all devices successfully",
		RevokedSessionCount: count,
	})
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

func (controller *AuthController) DeleteMe(ctx echo.Context) error {
	header, err := utils.GetContextHeaders(ctx, headers.ContextHeaders)
	if err != nil {
		return err
	}

	if err := controller.authService.DeleteAccount(ctx.Request().Context(), header.UserID); err != nil {
		return err
	}

	return response.SuccessWithMessage(ctx, "Account deleted successfully", &struct{}{})
}

func (controller *AuthController) ForgotPassword(ctx echo.Context) error {
	request := &dto.ForgotPasswordRequest{}

	if err := helper.BindAndValidate(ctx, request); err != nil {
		return err
	}

	if err := controller.authService.ForgotPassword(ctx.Request().Context(), request.Email); err != nil {
		return err
	}

	return response.SuccessWithMessage(ctx, "If the email is registered, an OTP will be sent", &struct{}{})
}

func (controller *AuthController) ResetPassword(ctx echo.Context) error {
	request := &dto.ResetPasswordRequest{}

	if err := helper.BindAndValidate(ctx, request); err != nil {
		return err
	}

	if err := controller.authService.ResetPassword(ctx.Request().Context(), request); err != nil {
		return err
	}

	return response.SuccessWithMessage(ctx, "Password reset successfully", &struct{}{})
}
