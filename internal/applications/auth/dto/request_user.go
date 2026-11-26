package dto

import "github.com/letenk/golang-authentication/exceptions"

type UpdateProfileRequest struct {
	// Optional, 3 to 100 characters. Send full_name, phone, or both.
	FullName string `json:"full_name" validate:"omitempty,min=3,max=100" example:"John Doe"`
	// Optional, E.164 format. Send full_name, phone, or both.
	Phone string `json:"phone" validate:"omitempty,e164" example:"+628123456789"`
}

// Validate custom validation: at least one field must be provided
func (r UpdateProfileRequest) Validate() error {
	if r.FullName == "" && r.Phone == "" {
		return exceptions.NewValidationError("full_name or phone", "at least one of full_name or phone is required")
	}
	return nil
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required" example:"Str0ng!Pass"`
	// At least 8 characters with a lowercase letter, an uppercase letter, a digit and a symbol.
	NewPassword string `json:"new_password" validate:"required,min=8,strong_password" example:"N3w!Passw0rd"`
}
