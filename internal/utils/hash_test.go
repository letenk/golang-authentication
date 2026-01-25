package utils_test

import (
	"testing"

	"github.com/letenk/golang-authentication/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{
			name:        "valid password",
			password:    "mySecret123!",
			expectError: false,
		},
		{
			name:        "empty password",
			password:    "",
			expectError: false, // bcrypt allows empty passwords
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashed, err := utils.HashPassword(tt.password)

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if hashed == "" {
				t.Fatalf("expected hashed password, got empty string")
			}

			// Ensure hash is valid bcrypt hash
			if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(tt.password)); err != nil {
				t.Fatalf("hash does not match password: %v", err)
			}
		})
	}
}

func TestComparePassword_TableDriven(t *testing.T) {
	password := "MyStrongPassword!"
	hashed, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password for test setup: %v", err)
	}

	tests := []struct {
		name        string
		hashed      string
		plain       string
		expectError bool
	}{
		{
			name:        "correct password",
			hashed:      hashed,
			plain:       password,
			expectError: false,
		},
		{
			name:        "wrong password",
			hashed:      hashed,
			plain:       "wrongPassword",
			expectError: true,
		},
		{
			name:        "invalid hash format",
			hashed:      "not-a-bcrypt-hash",
			plain:       password,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := utils.ComparePassword(tt.hashed, tt.plain)

			if tt.expectError && err == nil {
				t.Fatalf("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestHashPassword_UniqueHashes(t *testing.T) {
	password := "samePassword"

	hash1, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("hash1 failed: %v", err)
	}

	hash2, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("hash2 failed: %v", err)
	}

	if hash1 == hash2 {
		t.Fatalf("expected different hashes for same password")
	}
}
