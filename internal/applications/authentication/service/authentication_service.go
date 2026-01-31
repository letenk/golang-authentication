package service

import "github.com/golang-jwt/jwt/v5"

type AuthenticationService interface {
	ClaimToken(line string) (*jwt.MapClaims, error)
	ClaimUser(line string) (*jwt.MapClaims, *int64, error)
}