package jwt_config

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/letenk/golang-authentication/configs/credential"
)

type JWTConfig struct {
	secret             string
	accessTokenExpire  time.Duration
	refreshTokenExpire time.Duration
}

type JWTClaims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// func NewJWTConfig(secret string, accessTokenExpire, refreshTokenExpire string) (*JWTConfig, error) {
func NewJWTConfig(cfg *credential.JWTConfig) (*JWTConfig, error) {
	// Parse access token expiration
	accessDuration, err := parseDuration(cfg.AccessTokenExpire)
	if err != nil {
		return nil, fmt.Errorf("invalid access token expire format: %w", err)
	}

	// Parse refresh token expiration
	refreshDuration, err := parseDuration(cfg.RefreshTokenExpire)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token expire format: %w", err)
	}

	return &JWTConfig{
		secret:             cfg.Secret,
		accessTokenExpire:  accessDuration,
		refreshTokenExpire: refreshDuration,
	}, nil
}

// parseDuration parses duration strings like "15m", "24h", "7d"
func parseDuration(durationStr string) (time.Duration, error) {
	if len(durationStr) < 2 {
		return 0, fmt.Errorf("invalid duration format")
	}

	// Check if it ends with 'd' for days
	if durationStr[len(durationStr)-1] == 'd' {
		days := durationStr[:len(durationStr)-1]
		var numDays int
		_, err := fmt.Sscanf(days, "%d", &numDays)
		if err != nil {
			return 0, err
		}
		return time.Duration(numDays) * 24 * time.Hour, nil
	}

	// Otherwise use standard time.ParseDuration
	return time.ParseDuration(durationStr)
}

// GenerateAccessToken generates a new JWT access token
func (h *JWTConfig) GenerateAccessToken(userID int64) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(h.accessTokenExpire)

	claims := JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.secret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// GenerateRefreshToken generates a new UUID-based refresh token
func (h *JWTConfig) GenerateRefreshToken() (string, time.Time) {
	token := uuid.New().String()
	expiresAt := time.Now().UTC().Add(h.refreshTokenExpire)
	return token, expiresAt
}

// ValidateAccessToken validates and parses a JWT access token
func (h *JWTConfig) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GetAccessTokenExpire returns access token expiration duration
func (h *JWTConfig) GetAccessTokenExpire() time.Duration {
	return h.accessTokenExpire
}

// GetRefreshTokenExpire returns refresh token expiration duration
func (h *JWTConfig) GetRefreshTokenExpire() time.Duration {
	return h.refreshTokenExpire
}
