package database

import "database/sql"

// NullString stores "" as NULL. Kept here so every repository maps empty
// optional text the same way.
func NullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
