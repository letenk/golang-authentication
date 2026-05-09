package utils

import (
	"github.com/letenk/golang-authentication/exceptions"
	"github.com/nyaruka/phonenumbers"
)

// NormalizePhoneNumber parses an international phone number and returns it in
// E.164 format (e.g. "+628123456789"). The input must include a country code
// prefix with a leading "+" — there is no default region, so callers in any
// locale must provide a fully-qualified international number.
//
// Returns a ValidationError when the input is empty, missing the country code,
// or fails libphonenumber's validity check.
func NormalizePhoneNumber(number string) (string, error) {
	if number == "" {
		return "", exceptions.NewValidationError("phone", "phone number is required")
	}

	parsed, err := phonenumbers.Parse(number, "")
	if err != nil {
		return "", exceptions.NewValidationError("phone", "invalid phone number format (must include country code, e.g. +62...)")
	}

	if !phonenumbers.IsValidNumber(parsed) {
		return "", exceptions.NewValidationError("phone", "invalid phone number")
	}

	return phonenumbers.Format(parsed, phonenumbers.E164), nil
}
