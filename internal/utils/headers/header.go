package headers

import "github.com/golang-jwt/jwt/v5"

// Header represents HTTP headers and authentication data extracted from request
// This struct is populated by authentication middleware and stored in context
type Header struct {
	Authorization *string `header:"authorization" validate:"ascii"`
	UserID        int64
	IsUserGuest   bool `header:"x-user-guest" validate:"ascii"`
	Claim         *jwt.MapClaims
}
