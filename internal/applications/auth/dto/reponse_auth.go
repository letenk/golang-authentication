package dto

import "time"

// UserResponse represents user data in response
type UserResponse struct {
	ID         int64     `json:"id"`
	Email      string    `json:"email"`
	FullName   *string   `json:"full_name"`
	Phone      *string   `json:"phone"`
	IsVerified bool      `json:"is_verified"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}

// RegisterResponse represents registration response
type RegisterResponse struct {
	User    *UserResponse `json:"user"`
	Message string        `json:"message"`
}

// LoginResponse represents login response
type LoginResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresIn    int64         `json:"expires_in"` // seconds
}
