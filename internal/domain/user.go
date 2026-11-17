// Package domain holds the application's own entities and errors.
//
// Services and controllers depend only on these types, never on ORM or
// generated database types, so the persistence layer can change without
// touching business logic. This package imports only the standard library.
package domain

import "time"

// User is an application user. Nullable text columns are exposed as "" when NULL.
type User struct {
	ID         int64
	Name       string
	Email      string
	Phone      string
	Password   string // bcrypt hash
	LoginType  string // "email" | "phone"
	IsVerified bool
	VerifiedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
