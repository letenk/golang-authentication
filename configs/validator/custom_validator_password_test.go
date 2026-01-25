package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidPassword(t *testing.T) {
	validator := NewCustomValidator()

	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{
			name:     "Valid password with all requirements",
			password: "Password123!",
			valid:    true,
		},
		{
			name:     "Valid password with special chars",
			password: "SecureP@ss1",
			valid:    true,
		},
		{
			name:     "Missing uppercase",
			password: "password123!",
			valid:    false,
		},
		{
			name:     "Missing lowercase",
			password: "PASSWORD123!",
			valid:    false,
		},
		{
			name:     "Missing digit",
			password: "Password!",
			valid:    false,
		},
		{
			name:     "Missing special character",
			password: "Password123",
			valid:    false,
		},
		{
			name:     "Too short but meets other requirements",
			password: "Pass1!",
			valid:    false,
		},
		{
			name:     "Empty password",
			password: "",
			valid:    false,
		},
		{
			name:     "Only letters",
			password: "OnlyLetters",
			valid:    false,
		},
		{
			name:     "Only digits",
			password: "12345678",
			valid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type testStruct struct {
				Password string `validate:"required,min=8,strong_password"`
			}

			testData := testStruct{Password: tt.password}
			err := validator.Validate(testData)

			if tt.valid {
				assert.NoError(t, err, "Expected valid password: %s", tt.password)
			} else {
				assert.Error(t, err, "Expected invalid password: %s", tt.password)
			}
		})
	}
}