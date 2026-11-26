package controller

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/letenk/golang-authentication/internal/applications/auth/dto"
	"github.com/letenk/golang-authentication/internal/applications/auth/service"
	"github.com/letenk/golang-authentication/internal/helper"
	"github.com/letenk/golang-authentication/internal/helper/response"
	"github.com/letenk/golang-authentication/internal/utils"
	"github.com/letenk/golang-authentication/internal/utils/headers"
)

type UserController struct {
	authService service.AuthService
}

func NewUserController(authService service.AuthService) *UserController {
	return &UserController{authService: authService}
}

// GetSessions returns all active sessions for the authenticated user
//
//	@Summary		List active sessions
//	@Tags			user
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.body{data=[]dto.ActiveSessionResponse}
//	@Failure		401	{object}	response.body	"Missing, invalid or expired access token"
//	@Failure		422	{object}	response.body	"Failed to load sessions"
//	@Failure		500	{object}	response.body	"Unexpected server error"
//	@Router			/user/sessions [get]
func (controller *UserController) GetSessions(ctx echo.Context) error {
	header, err := utils.GetContextHeaders(ctx, headers.ContextHeaders)
	if err != nil {
		return err
	}

	sessions, err := controller.authService.GetSessions(ctx.Request().Context(), header.UserID)
	if err != nil {
		return err
	}

	return response.SuccessWithMessage(ctx, "Sessions retrieved successfully", sessions)
}

// RevokeSession revokes a specific session by ID
//
//	@Summary		Revoke a session
//	@Tags			user
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Session ID (from GET /user/sessions)"
//	@Success		200	{object}	response.body
//	@Failure		400	{object}	response.body	"Session ID is not a number"
//	@Failure		401	{object}	response.body	"Missing, invalid or expired access token"
//	@Failure		404	{object}	response.body	"Session not found for this user"
//	@Failure		500	{object}	response.body	"Unexpected server error"
//	@Router			/user/sessions/{id} [delete]
func (controller *UserController) RevokeSession(ctx echo.Context) error {
	header, err := utils.GetContextHeaders(ctx, headers.ContextHeaders)
	if err != nil {
		return err
	}

	idParam := ctx.Param("id")
	sessionID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid session id")
	}

	if err := controller.authService.RevokeSession(ctx.Request().Context(), header.UserID, sessionID); err != nil {
		return err
	}

	return response.SuccessWithMessage(ctx, "Session revoked successfully", &struct{}{})
}

// UpdateProfile updates the authenticated user's full name and/or phone number
//
//	@Summary		Update the profile
//	@Description	Updates full name and/or phone (E.164, e.g. +628123456789).
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body	dto.UpdateProfileRequest	true	"Fields to change"
//	@Success		200	{object}	response.body{data=dto.UserResponse}
//	@Failure		400	{object}	response.body	"Validation failed; details in errors"
//	@Failure		401	{object}	response.body	"Missing, invalid or expired access token"
//	@Failure		404	{object}	response.body	"User not found"
//	@Failure		422	{object}	response.body	"Phone already in use"
//	@Failure		500	{object}	response.body	"Unexpected server error"
//	@Router			/user/profile [put]
func (controller *UserController) UpdateProfile(ctx echo.Context) error {
	header, err := utils.GetContextHeaders(ctx, headers.ContextHeaders)
	if err != nil {
		return err
	}

	req := &dto.UpdateProfileRequest{}
	if err := helper.BindAndValidate(ctx, req); err != nil {
		return err
	}

	if err := req.Validate(); err != nil {
		return err
	}

	result, err := controller.authService.UpdateProfile(ctx.Request().Context(), header.UserID, req)
	if err != nil {
		return err
	}

	return response.SuccessWithMessage(ctx, "Profile updated successfully", result)
}

// UpdatePassword changes the authenticated user's password
//
//	@Summary		Change the password
//	@Description	The new password needs at least 8 characters and must pass the strong-password rule.
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body	dto.UpdatePasswordRequest	true	"Old and new password"
//	@Success		200	{object}	response.body
//	@Failure		400	{object}	response.body	"Validation failed; details in errors"
//	@Failure		401	{object}	response.body	"Missing, invalid or expired access token"
//	@Failure		404	{object}	response.body	"User not found"
//	@Failure		422	{object}	response.body	"Old password is incorrect"
//	@Failure		500	{object}	response.body	"Unexpected server error"
//	@Router			/user/password [put]
func (controller *UserController) UpdatePassword(ctx echo.Context) error {
	header, err := utils.GetContextHeaders(ctx, headers.ContextHeaders)
	if err != nil {
		return err
	}

	req := &dto.UpdatePasswordRequest{}
	if err := helper.BindAndValidate(ctx, req); err != nil {
		return err
	}

	if err := controller.authService.UpdatePassword(ctx.Request().Context(), header.UserID, req); err != nil {
		return err
	}

	return response.SuccessWithMessage(ctx, "Password updated successfully", &struct{}{})
}
