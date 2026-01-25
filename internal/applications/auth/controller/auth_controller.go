package controller

import (
	"github.com/labstack/echo/v4"
	"github.com/letenk/golang-authentication/internal/applications/auth/dto"
	"github.com/letenk/golang-authentication/internal/applications/auth/service"
	"github.com/letenk/golang-authentication/internal/helper"
	"github.com/letenk/golang-authentication/internal/helper/response"
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

// func (controller *AuthController) RefreshToken(ctx echo.Context) error {
// 	request := &dto.RefreshTokenRequest{}

// 	err := helper.BindAndValidate(ctx, request)
// 	if err != nil {
// 		return err
// 	}

// 	result, err := controller.authService.RefreshToken(ctx.Request().Context(), request)
// 	if err != nil {
// 		return err
// 	}

// 	return response.SuccessWithMessage(ctx, "Token refreshed successfully", result)
// }

// func (controller *AuthController) Logout(ctx echo.Context) error {
// 	request := &dto.LogoutRequest{}

// 	err := helper.BindAndValidate(ctx, request)
// 	if err != nil {
// 		return err
// 	}

// 	// Get user ID from JWT middleware context
// 	userID := ctx.Get("user_id").(int64)

// 	err = controller.authService.Logout(ctx.Request().Context(), request, userID)
// 	if err != nil {
// 		return err
// 	}

// 	return response.SuccessWithMessage(ctx, "Logged out successfully", nil)
// }

// func (controller *AuthController) LogoutAll(ctx echo.Context) error {
// 	// Get user ID from JWT middleware context
// 	userID := ctx.Get("user_id").(int64)

// 	result, err := controller.authService.LogoutAll(ctx.Request().Context(), userID)
// 	if err != nil {
// 		return err
// 	}

// 	return response.SuccessWithMessage(ctx, result.Message, map[string]int{
// 		"revoked_sessions_count": result.RevokedSessionCount,
// 	})
// }

// func (controller *AuthController) GetMe(ctx echo.Context) error {
// 	// Get user ID from JWT middleware context
// 	userID := ctx.Get("user_id").(int64)

// 	result, err := controller.authService.GetMe(ctx.Request().Context(), userID)
// 	if err != nil {
// 		return err
// 	}

// 	return response.SuccessWithMessage(ctx, "User retrieved successfully", result)
// }
