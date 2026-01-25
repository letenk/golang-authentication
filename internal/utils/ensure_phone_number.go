package utils

import (
	"regexp"
	"strings"
)

// EnsurePhoneNumber ensures the phone number has Indonesian country code format (62)
// Examples:
//   - "08123456789" -> "628123456789"
//   - "+628123456789" -> "628123456789"
//   - "628123456789" -> "628123456789"
//   - "008123456789" -> "628123456789"
func EnsurePhoneNumber(number string) string {
	// If already starts with 628, return as is
	if strings.HasPrefix(number, "628") {
		return number
	}

	// Remove prefix +62, 62, or 0 from the beginning of the number
	re := regexp.MustCompile(`^[+620]+`)
	cleaned := re.ReplaceAllString(number, "")

	// Add prefix 62
	return "62" + cleaned
}