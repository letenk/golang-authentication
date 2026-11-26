package dto

type RegisterRequest struct {
	// Email or phone is required; both may be set.
	Email string `json:"email" validate:"omitempty,email" example:"user@example.com"`
	// E.164 format. Email or phone is required; both may be set.
	Phone string `json:"phone" validate:"omitempty,e164" example:"+628123456789"`
	// At least 8 characters with a lowercase letter, an uppercase letter, a digit and a symbol.
	Password string `json:"password" validate:"required,min=8,strong_password" example:"Str0ng!Pass"`
	// Optional, 3 to 100 characters.
	FullName string `json:"full_name" validate:"omitempty,min=3,max=100" example:"John Doe"`
}

type LoginRequest struct {
	// Email or phone is required. If both are sent, email is used.
	Email string `json:"email" validate:"omitempty,email" example:"user@example.com"`
	// E.164 format. Used only when email is empty.
	Phone    string `json:"phone" validate:"omitempty,e164" example:"+628123456789"`
	Password string `json:"password" validate:"required" example:"Str0ng!Pass"`
	// Optional. Label shown in GET /user/sessions; defaults to "Unknown Device".
	DeviceName string `json:"device_name" validate:"omitempty,max=100" example:"iPhone 15"`
	// Optional. Client-generated identifier of the device.
	DeviceID string `json:"device_id" validate:"omitempty,max=255" example:"a1b2c3d4"`
}

type RefreshTokenRequest struct {
	// Refresh token from login or the previous refresh. Leave the body empty to use the refresh_token cookie.
	RefreshToken string `json:"refresh_token" validate:"required,uuid" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
}

type LogoutRequest struct {
	// Refresh token of the session to end. Leave the body empty to use the refresh_token cookie.
	RefreshToken string `json:"refresh_token" validate:"required,uuid" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email" example:"user@example.com"`
}

type ResetPasswordRequest struct {
	Email string `json:"email" validate:"required,email" example:"user@example.com"`
	// OTP from the password-reset email.
	OTP string `json:"otp" validate:"required" example:"12345"`
	// At least 8 characters with a lowercase letter, an uppercase letter, a digit and a symbol.
	NewPassword string `json:"new_password" validate:"required,min=8,strong_password" example:"N3w!Passw0rd"`
}

type VerifyEmailRequest struct {
	// OTP from the verification email.
	OTP string `json:"otp" validate:"required" example:"12345"`
}
