package controller

import (
	"github.com/labstack/echo/v4"
	"github.com/letenk/golang-authentication/exceptions"
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

// Register creates a new account.
//
//	@Summary		Register a new user
//	@Description	Either email or phone (E.164, e.g. +628123456789) is required. The password needs at least 8 characters and must pass the strong-password rule.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.RegisterRequest	true	"Registration data"
//	@Success		200	{object}	response.body{data=dto.RegisterResponse}
//	@Failure		400	{object}	response.body	"Validation failed; details in errors"
//	@Failure		422	{object}	response.body	"Email or phone already registered"
//	@Failure		500	{object}	response.body	"Unexpected server error"
//	@Router			/auth/register [post]
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
			FullName:   result.Name,
			Email:      result.Email,
			Phone:      result.Phone,
			IsVerified: result.IsVerified,
		},
	}

	return response.SuccessWithMessage(ctx, "User registered successfully", successResponse)
}

// Login authenticates with email or phone and starts a session.
//
//	@Summary		Log in
//	@Description	Log in with email or phone plus password. Returns the tokens in the body and also sets the access_token and refresh_token HttpOnly cookies. Rate limited to 5 requests per minute per IP.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.LoginRequest	true	"Credentials and optional device info"
//	@Success		200	{object}	response.body{data=dto.LoginResponse}
//	@Failure		400	{object}	response.body	"Validation failed; details in errors"
//	@Failure		401	{object}	response.body	"Invalid credentials"
//	@Failure		429	{object}	response.body	"Rate limit exceeded"
//	@Failure		500	{object}	response.body	"Unexpected server error"
//	@Router			/auth/login [post]
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

	helper.SetAuthCookies(ctx, result.AccessToken, result.RefreshToken)

	return response.SuccessWithMessage(ctx, "Login successful", result)
}

// RefreshToken rotates the refresh token and issues a new access token.
//
//	@Summary		Refresh tokens
//	@Description	The refresh token is read from the body, or from the refresh_token cookie when the body is empty. The old refresh token is revoked (rotation) and new cookies are set. Rate limited to 10 requests per minute per IP.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.RefreshTokenRequest	false	"Optional when the refresh_token cookie is sent"
//	@Success		200	{object}	response.body{data=dto.RefreshTokenResponse}
//	@Failure		401	{object}	response.body	"Refresh token missing, invalid, revoked or expired"
//	@Failure		429	{object}	response.body	"Rate limit exceeded"
//	@Failure		500	{object}	response.body	"Unexpected server error"
//	@Router			/auth/refresh [post]
func (controller *AuthController) RefreshToken(ctx echo.Context) error {
	request := &dto.RefreshTokenRequest{}

	// Body is optional: the refresh token may arrive as an HttpOnly cookie
	// when the web client can't read it from JS to put it in the body.
	helper.BindAndValidate(ctx, request)

	if request.RefreshToken == "" {
		request.RefreshToken = helper.GetRefreshTokenFromCookie(ctx)
	}
	if request.RefreshToken == "" {
		return exceptions.NewAuthenticationError("refresh token is required", exceptions.AuthenticationInvalidToken)
	}

	ipAddress := ctx.RealIP()
	userAgent := ctx.Request().UserAgent()

	result, err := controller.authService.RefreshToken(ctx.Request().Context(), request, ipAddress, userAgent)
	if err != nil {
		return err
	}

	helper.SetAuthCookies(ctx, result.AccessToken, result.RefreshToken)

	return response.SuccessWithMessage(ctx, "Token refreshed successfully", result)
}

// Logout revokes the current session's refresh token.
//
//	@Summary		Log out the current session
//	@Description	Revokes the given refresh token and clears the auth cookies. No access token is needed: the refresh token is read from the body, or from the refresh_token cookie when the body is empty.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.LogoutRequest	false	"Optional when the refresh_token cookie is sent"
//	@Success		200	{object}	response.body{data=dto.LogoutResponse}
//	@Failure		401	{object}	response.body	"Refresh token missing, invalid or already revoked"
//	@Failure		500	{object}	response.body	"Unexpected server error"
//	@Router			/auth/logout [post]
func (controller *AuthController) Logout(ctx echo.Context) error {
	request := &dto.LogoutRequest{}

	// Body is optional: the refresh token may arrive as an HttpOnly cookie.
	helper.BindAndValidate(ctx, request)

	refreshToken := request.RefreshToken
	if refreshToken == "" {
		refreshToken = helper.GetRefreshTokenFromCookie(ctx)
	}
	if refreshToken == "" {
		return exceptions.NewAuthenticationError("refresh token is required", exceptions.AuthenticationInvalidToken)
	}

	err := controller.authService.Logout(ctx.Request().Context(), refreshToken)
	if err != nil {
		return err
	}

	helper.ClearAuthCookies(ctx)

	return response.SuccessWithMessage(ctx, "Logged out successfully", &dto.LogoutResponse{
		Message: "Logged out successfully",
	})
}

// LogoutAll revokes every session of the authenticated user.
//
//	@Summary		Log out all devices
//	@Description	Revokes all refresh tokens of the current user and clears the auth cookies.
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.body{data=dto.LogoutAllResponse}
//	@Failure		401	{object}	response.body	"Missing, invalid or expired access token"
//	@Failure		422	{object}	response.body	"Failed to revoke sessions"
//	@Failure		500	{object}	response.body	"Unexpected server error"
//	@Router			/auth/logout-all [post]
func (controller *AuthController) LogoutAll(ctx echo.Context) error {
	header, err := utils.GetContextHeaders(ctx, headers.ContextHeaders)
	if err != nil {
		return err
	}

	count, err := controller.authService.LogoutAll(ctx.Request().Context(), header.UserID)
	if err != nil {
		return err
	}

	helper.ClearAuthCookies(ctx)

	return response.SuccessWithMessage(ctx, "Logged out from all devices successfully", &dto.LogoutAllResponse{
		Message:             "Logged out from all devices successfully",
		RevokedSessionCount: count,
	})
}

// GetMe returns the authenticated user and their active sessions.
//
//	@Summary		Get the current user
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.body{data=dto.GetMeResponse}
//	@Failure		401	{object}	response.body	"Missing, invalid or expired access token"
//	@Failure		404	{object}	response.body	"User not found"
//	@Failure		500	{object}	response.body	"Unexpected server error"
//	@Router			/auth/me [get]
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

// DeleteMe deletes the authenticated user's account.
//
//	@Summary		Delete the current account
//	@Description	Soft-deletes the account of the current user.
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.body
//	@Failure		401	{object}	response.body	"Missing, invalid or expired access token"
//	@Failure		422	{object}	response.body	"Failed to delete the account"
//	@Failure		500	{object}	response.body	"Unexpected server error"
//	@Router			/auth/me [delete]
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

// ForgotPassword sends a password-reset OTP.
//
//	@Summary		Request a password-reset OTP
//	@Description	Always answers 200 for a well-formed email, whether or not it is registered, so the endpoint cannot be used to discover accounts. If the email is registered, an OTP is sent to it.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.ForgotPasswordRequest	true	"Account email"
//	@Success		200	{object}	response.body
//	@Failure		400	{object}	response.body	"Validation failed; details in errors"
//	@Failure		500	{object}	response.body	"Unexpected server error"
//	@Router			/auth/forgot-password [post]
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

// ResetPassword sets a new password using the emailed OTP.
//
//	@Summary		Reset the password with an OTP
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.ResetPasswordRequest	true	"Email, OTP and new password"
//	@Success		200	{object}	response.body
//	@Failure		400	{object}	response.body	"Validation failed; details in errors"
//	@Failure		404	{object}	response.body	"Invalid email or OTP"
//	@Failure		422	{object}	response.body	"Invalid or expired OTP"
//	@Failure		500	{object}	response.body	"Unexpected server error"
//	@Router			/auth/reset-password [post]
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

// VerifyEmail confirms the authenticated user's email with an OTP.
//
//	@Summary		Verify the email address
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body	dto.VerifyEmailRequest	true	"OTP from the verification email"
//	@Success		200	{object}	response.body
//	@Failure		400	{object}	response.body	"Validation failed; details in errors"
//	@Failure		401	{object}	response.body	"Missing, invalid or expired access token"
//	@Failure		422	{object}	response.body	"Invalid or expired OTP"
//	@Failure		500	{object}	response.body	"Unexpected server error"
//	@Router			/auth/verify-email [post]
func (controller *AuthController) VerifyEmail(ctx echo.Context) error {
	header, err := utils.GetContextHeaders(ctx, headers.ContextHeaders)
	if err != nil {
		return err
	}

	request := &dto.VerifyEmailRequest{}
	if err := helper.BindAndValidate(ctx, request); err != nil {
		return err
	}

	if err := controller.authService.VerifyEmail(ctx.Request().Context(), header.UserID, request.OTP); err != nil {
		return err
	}

	return response.SuccessWithMessage(ctx, "Email verified successfully", &struct{}{})
}

// ResendVerificationEmail sends a new email-verification OTP.
//
//	@Summary		Resend the verification email
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.body
//	@Failure		401	{object}	response.body	"Missing, invalid or expired access token"
//	@Failure		404	{object}	response.body	"User not found"
//	@Failure		422	{object}	response.body	"No email address, or the email is already verified"
//	@Failure		500	{object}	response.body	"Unexpected server error"
//	@Router			/auth/resend-verification-email [post]
func (controller *AuthController) ResendVerificationEmail(ctx echo.Context) error {
	header, err := utils.GetContextHeaders(ctx, headers.ContextHeaders)
	if err != nil {
		return err
	}

	if err := controller.authService.SendVerificationEmail(ctx.Request().Context(), header.UserID); err != nil {
		return err
	}

	return response.SuccessWithMessage(ctx, "Verification email sent", &struct{}{})
}
