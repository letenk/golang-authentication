package middleware

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"

	"github.com/letenk/golang-authentication/exceptions"
	"github.com/letenk/golang-authentication/internal/applications/authentication/service"
	"github.com/letenk/golang-authentication/internal/utils/headers"
)

type AuthenticationMiddlewareImpl struct {
	service service.AuthenticationService
}

func NewAuthenticationMiddleware(service service.AuthenticationService) *AuthenticationMiddlewareImpl {
	return &AuthenticationMiddlewareImpl{
		service: service,
	}
}

// Authenticate validates JWT token and extracts user claims
// If required is true, returns error when no token is provided
// If required is false, proceeds without token
func (impl *AuthenticationMiddlewareImpl) Authenticate(required bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			header, err := bindHeader(ctx)
			if err != nil {
				return err
			}

			if header.Authorization != nil {
				claim, userId, err := impl.service.ClaimUser(*header.Authorization)
				if err != nil {
					err := exceptions.NewAuthenticationError(
						err.Error(),
						exceptions.AuthenticationBadFormat,
					)
					log.Errorf("failed to claim user - invalid Authorization token: %s", err)
					return err
				}

				buildHeader(header, *userId, claim)
				requestCtx := ctx.Request().Context()
				requestCtx = context.WithValue(requestCtx, headers.ContextHeaders, header)
				ctx.SetRequest(ctx.Request().WithContext(requestCtx))
			} else if required {
				err := exceptions.NewAuthenticationError(
					"Unauthorized - No token provided",
					exceptions.AuthenticationUnauthenticated,
				)
				log.Errorf("failed to claim user - no Authorization token provided: %s", err)
				return err
			}

			return next(ctx)
		}
	}
}

// bindHeader binds HTTP headers to Header struct
func bindHeader(ctx echo.Context) (*headers.Header, error) {
	header := &headers.Header{}

	if err := (&echo.DefaultBinder{}).BindHeaders(ctx, header); err != nil {
		return nil, err
	}

	return header, nil
}

// buildHeader populates header with user authentication data
func buildHeader(header *headers.Header, userId int64, claim *jwt.MapClaims) *headers.Header {
	header.UserID = userId
	header.Claim = claim
	header.IsUserGuest = false
	return header
}
