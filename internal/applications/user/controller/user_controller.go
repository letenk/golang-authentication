package controller

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/letenk/golang-authentication/internal/applications/auth/service"
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
