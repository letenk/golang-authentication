package dto

type RegisterRequest struct {
	Email    string `json:"email" validate:"omitempty,email"`
	Phone    string `json:"phone" validate:"omitempty,e164"` // E.164 format: +628123456789
	Password string `json:"password" validate:"required,min=8,strong_password"`
	FullName string `json:"full_name" validate:"omitempty,min=3,max=100"`
}

type LoginRequest struct {
	Email      string `json:"email" validate:"omitempty,email"`
	Phone      string `json:"phone" validate:"omitempty,e164"` // E.164 format: +628123456789
	Password   string `json:"password" validate:"required"`
	DeviceName string `json:"device_name" validate:"omitempty,max=100"`
	DeviceID   string `json:"device_id" validate:"omitempty,max=255"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,uuid"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,uuid"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" validate:"required,email"`
	OTP         string `json:"otp" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,strong_password"`
}
