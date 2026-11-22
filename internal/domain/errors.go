package domain

import "errors"

var (
	// ErrNotFound is returned by repositories when no row matches, or when an
	// update/delete affects no row.
	ErrNotFound = errors.New("record not found")
	// ErrEmailTaken is returned when the email unique constraint is violated.
	ErrEmailTaken = errors.New("email already registered")
	// ErrPhoneTaken is returned when the phone unique constraint is violated.
	ErrPhoneTaken = errors.New("phone already registered")
)
