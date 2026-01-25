package dto

type RegisterRequest struct {
	Email    string `json:"email" validate:"omitempty,email"`
	Phone    string `json:"phone" validate:"omitempty,e164"` // E.164 format: +628123456789
	Password string `json:"password" validate:"required,min=8,strong_password"`
	FullName string `json:"full_name" validate:"omitempty,min=3,max=100"`
}