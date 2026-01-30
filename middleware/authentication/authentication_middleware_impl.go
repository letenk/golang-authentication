package middleware

import (
	"context"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"

	jwt_config "github.com/letenk/golang-authentication/configs/jwt_config"
	"github.com/letenk/golang-authentication/exceptions"
)

const (
	ContextKeyUserID = "user_id"
	ContextKeyEmail  = "email"
	ContextKeyClaims = "claims"
)

type AuthenticationMiddlewareImpl struct {
	jwtConfig *jwt_config.JWTConfig
}

func NewAuthenticationMiddleware(jwtConfig *jwt_config.JWTConfig) *AuthenticationMiddlewareImpl {
	return &AuthenticationMiddlewareImpl{
		jwtConfig: jwtConfig,
	}
}

// Authenticate validates JWT token and extracts user claims
// If required is true, returns error when no token is provided
// If required is false, proceeds without token
func (m *AuthenticationMiddlewareImpl) Authenticate(required bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract token from Authorization header
			authHeader := c.Request().Header.Get("Authorization")

			if authHeader == "" {
				if required {
					err := exceptions.NewAuthenticationError(
						"Unauthorized - No token provided",
						exceptions.AuthenticationUnauthenticated,
					)
					log.Error("No authorization header provided")
					return err
				}
				return next(c)
			}

			// Validate Bearer format
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				if required {
					err := exceptions.NewAuthenticationError(
						"Invalid authorization header format",
						exceptions.AuthenticationBadFormat,
					)
					log.Error("Invalid Bearer token format")
					return err
				}
				return next(c)
			}

			tokenString := parts[1]

			// Validate JWT token
			claims, err := m.jwtConfig.ValidateAccessToken(tokenString)
			if err != nil {
				if required {
					err := exceptions.NewAuthenticationError(
						"Invalid or expired access token",
						exceptions.AuthenticationInvalidToken,
					)
					log.Errorf("JWT validation failed: %v", err)
					return err
				}
				return next(c)
			}

			// Add claims to echo context (Echo's way)
			c.Set(ContextKeyUserID, claims.UserID)
			c.Set(ContextKeyEmail, claims.Email)
			c.Set(ContextKeyClaims, claims)

			// Also add to request context for services
			ctx := c.Request().Context()
			ctx = context.WithValue(ctx, ContextKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, ContextKeyEmail, claims.Email)
			ctx = context.WithValue(ctx, ContextKeyClaims, claims)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

// ExtractUserID retrieves user ID from context
func ExtractUserID(c echo.Context) (int64, error) {
	userID := c.Get(ContextKeyUserID)
	if userID == nil {
		return 0, exceptions.NewAuthenticationError(
			"User ID not found in context",
			exceptions.AuthenticationUnauthenticated,
		)
	}
	return userID.(int64), nil
}

// ExtractEmail retrieves email from context
func ExtractEmail(c echo.Context) (string, error) {
	email := c.Get(ContextKeyEmail)
	if email == nil {
		return "", exceptions.NewAuthenticationError(
			"Email not found in context",
			exceptions.AuthenticationUnauthenticated,
		)
	}
	return email.(string), nil
}
