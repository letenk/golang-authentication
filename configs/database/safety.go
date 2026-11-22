package database

import (
	"context"
	"database/sql"
	"fmt"
)

// RowQuerier is satisfied by *sql.DB, *sql.Conn and *sql.Tx.
type RowQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// CheckStandardConformingStrings fails unless the server treats backslashes in
// string literals literally. Bun interpolates query values into the SQL text
// on the client and escapes strings only by doubling single quotes, which is
// injection-safe only while standard_conforming_strings is on (the PostgreSQL
// default since 9.1). A role or database configured with it off must not run
// this application.
func CheckStandardConformingStrings(ctx context.Context, q RowQuerier) error {
	var value string
	if err := q.QueryRowContext(ctx, "SHOW standard_conforming_strings").Scan(&value); err != nil {
		return fmt.Errorf("reading standard_conforming_strings: %w", err)
	}
	if value != "on" {
		return fmt.Errorf("standard_conforming_strings is %q; it must be \"on\" because query values are escaped client-side", value)
	}
	return nil
}
