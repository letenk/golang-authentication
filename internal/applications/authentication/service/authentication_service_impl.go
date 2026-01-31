package service

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gookit/goutil/arrutil"
	"github.com/letenk/golang-authentication/configs/jwt_config"
	authenticationError "github.com/letenk/golang-authentication/exceptions"
)

type AuthenticationServiceImpl struct {
	authentication *jwt_config.JWTConfig
}

var authenticationRealms = []string{"bearer", "jwt"}

func NewAuthenticationService(
	authentication *jwt_config.JWTConfig,
) *AuthenticationServiceImpl {
	return &AuthenticationServiceImpl{
		authentication: authentication,
	}
}

func unpackToken(token string) (*string, error) {
	parts := strings.SplitN(token, " ", 2)

	if len(parts) != 2 {
		return nil, authenticationError.NewAuthenticationError("invalid authentication format", authenticationError.AuthenticationBadFormat)
	}

	if arrutil.In(strings.ToLower(parts[0]), authenticationRealms) {
		return &parts[1], nil
	} else {
		return nil, authenticationError.NewAuthenticationError(fmt.Sprintf("%s is not valid realm type", parts[0]), authenticationError.AuthenticationInvalidRealm)
	}
}

func (authentication *AuthenticationServiceImpl) ClaimToken(line string) (*jwt.MapClaims, error) {
	token, err := unpackToken(line)
	if err != nil {
		return nil, err
	}

	// Use ValidateAccessToken from jwt_config
	claims, err := authentication.authentication.ValidateAccessToken(*token)
	if err != nil {
		return nil, err
	}

	// Convert JWTClaims to jwt.MapClaims
	mapClaims := jwt.MapClaims{
		"user_id": claims.UserID,
		"sub":     fmt.Sprintf("%d", claims.UserID),
		"exp":     claims.ExpiresAt.Unix(),
		"iat":     claims.IssuedAt.Unix(),
		"jti":     claims.ID,
	}

	return &mapClaims, nil
}

func (authentication *AuthenticationServiceImpl) ClaimUser(line string) (*jwt.MapClaims, *int64, error) {
	claim, err := authentication.ClaimToken(line)
	if err != nil {
		return nil, nil, err
	}

	subject, ok := (*claim)["sub"]
	if !ok {
		return nil, nil, authenticationError.NewAuthenticationError("subject is empty", authenticationError.AuthenticationInvalidIdentity)
	}

	userId, err := strconv.ParseInt(subject.(string), 10, 64)
	if err != nil {
		return nil, nil, authenticationError.NewAuthenticationError("invalid subject value", authenticationError.AuthenticationInvalidIdentity)
	}

	return claim, &userId, nil
}
