package dto

import "github.com/letenk/golang-authentication/exceptions"

type UpdateProfileRequest struct {
	FullName string `json:"full_name" validate:"omitempty,min=3,max=100"`
	Phone    string `json:"phone" validate:"omitempty,e164"`
}

// Validate custom validation: at least one field must be provided
func (r UpdateProfileRequest) Validate() error {
	if r.FullName == "" && r.Phone == "" {
		return exceptions.NewValidationError("full_name or phone", "at least one of full_name or phone is required")
	}
	return nil
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,strong_password"`
}
