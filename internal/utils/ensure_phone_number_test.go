package utils

import (
	"testing"
)

func TestEnsurePhoneNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Number already in 628 format",
			input:    "628123456789",
			expected: "628123456789",
		},
		{
			name:     "Number with prefix 0",
			input:    "08123456789",
			expected: "628123456789",
		},
		{
			name:     "Number with prefix +62",
			input:    "+628123456789",
			expected: "628123456789",
		},
		{
			name:     "Number with prefix 62",
			input:    "628987654321",
			expected: "628987654321",
		},
		{
			name:     "Number with prefix 00",
			input:    "008123456789",
			expected: "628123456789",
		},
		{
			name:     "Number with prefix +62 without 8",
			input:    "+62123456789",
			expected: "62123456789",
		},
		{
			name:     "Number with multiple leading zeros",
			input:    "0008123456789",
			expected: "628123456789",
		},
		{
			name:     "Number with spaces (not cleaned)",
			input:    "0812 3456 789",
			expected: "62812 3456 789",
		},
		{
			name:     "Empty number",
			input:    "",
			expected: "62",
		},
		{
			name:     "Number with only +62 prefix",
			input:    "+62",
			expected: "62",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EnsurePhoneNumber(tt.input)
			if result != tt.expected {
				t.Errorf("EnsurePhoneNumber(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}