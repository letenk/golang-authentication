package database

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const sqlStateUniqueViolation = "23505"

// UniqueViolationConstraint reports whether err is a PostgreSQL unique
// violation and, if so, the name of the violated constraint or index.
func UniqueViolationConstraint(err error) (string, bool) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == sqlStateUniqueViolation {
		return pgErr.ConstraintName, true
	}
	return "", false
}
