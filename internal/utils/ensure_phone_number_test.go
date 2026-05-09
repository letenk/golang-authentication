package utils

import (
	"testing"
)

func TestNormalizePhoneNumber_Valid(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Indonesia mobile",
			input:    "+628123456789",
			expected: "+628123456789",
		},
		{
			name:     "Indonesia mobile with spaces",
			input:    "+62 812 3456 789",
			expected: "+628123456789",
		},
		{
			name:     "Indonesia mobile with dashes",
			input:    "+62-812-3456-789",
			expected: "+628123456789",
		},
		{
			name:     "US number",
			input:    "+14155552671",
			expected: "+14155552671",
		},
		{
			name:     "US number with formatting",
			input:    "+1 (415) 555-2671",
			expected: "+14155552671",
		},
		{
			name:     "UK number",
			input:    "+447911123456",
			expected: "+447911123456",
		},
		{
			name:     "Japan number",
			input:    "+819012345678",
			expected: "+819012345678",
		},
		{
			name:     "Singapore number",
			input:    "+6591234567",
			expected: "+6591234567",
		},
		{
			name:     "Malaysia number",
			input:    "+60123456789",
			expected: "+60123456789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizePhoneNumber(tt.input)
			if err != nil {
				t.Fatalf("NormalizePhoneNumber(%q) returned error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("NormalizePhoneNumber(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizePhoneNumber_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "no country code prefix",
			input: "08123456789",
		},
		{
			name:  "missing plus sign",
			input: "628123456789",
		},
		{
			name:  "too short",
			input: "+123",
		},
		{
			name:  "non-numeric",
			input: "+abc12345",
		},
		{
			name:  "only plus sign",
			input: "+",
		},
		{
			name:  "invalid country code",
			input: "+9991234567890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizePhoneNumber(tt.input)
			if err == nil {
				t.Errorf("NormalizePhoneNumber(%q) expected error, got result %q", tt.input, result)
			}
		})
	}
}
