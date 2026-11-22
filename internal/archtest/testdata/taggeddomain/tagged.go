// Package taggeddomain is a fixture for internal/archtest: a "domain" struct
// carrying persistence tags. It lives under testdata/, so `go build` never
// compiles it.
package taggeddomain

type User struct {
	Email string `bun:"email"`
	Name  string `db:"name" json:"name"`
	Bio   string `json:"bio"`
	Plain string
}
