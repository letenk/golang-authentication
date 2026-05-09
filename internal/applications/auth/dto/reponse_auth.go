package dto

// UserResponse represents user data in response
type UserResponse struct {
	ID         int64  `json:"id"`
	Email      string `json:"email"`
	FullName   string `json:"full_name"`
	Phone      string `json:"phone"`
	IsVerified bool   `json:"is_verified"`
	IsActive   bool   `json:"is_active"`
}

// RegisterResponse represents registration response
type RegisterResponse struct {
	User *UserResponse `json:"user"`
}

// LoginResponse represents login response
type LoginResponse struct {
	AccessToken           string `json:"access_token"`
	RefreshToken          string `json:"refresh_token"`
	AccessTokenExpiresAt  string `json:"access_token_expires_at"`
	RefreshTokenExpiresAt string `json:"refresh_token_expires_at"`
}

// RefreshTokenResponse represents token refresh response
type RefreshTokenResponse struct {
	AccessToken            string `json:"access_token"`
	RefreshToken           string `json:"refresh_token"`
	AccessTokenExpiresAt   string `json:"access_token_expires_at"`
	RefreshTokenExpiresAt  string `json:"refresh_token_expires_at"`
}

// LogoutResponse represents logout response
type LogoutResponse struct {
	Message string `json:"message"`
}

// LogoutAllResponse represents logout all devices response
type LogoutAllResponse struct {
	Message             string `json:"message"`
	RevokedSessionCount int    `json:"revoked_sessions_count"`
}

// ActiveSessionResponse represents a single active session
type ActiveSessionResponse struct {
	ID         int64  `json:"id"`
	DeviceName string `json:"device_name"`
	DeviceID   string `json:"device_id"`
	IPAddress  string `json:"ip_address"`
	UserAgent  string `json:"user_agent"`
	LastUsed   string `json:"last_used"`
	CreatedAt  string `json:"created_at"`
}

// GetMeResponse represents get current user response
type GetMeResponse struct {
	User           *UserResponse            `json:"user"`
	ActiveSessions []*ActiveSessionResponse `json:"active_sessions"`
}
