package database_test

import (
	"database/sql"
	"testing"

	"github.com/letenk/golang-authentication/configs/database"
	"github.com/stretchr/testify/assert"
)

func TestNullString(t *testing.T) {
	assert.Equal(t, sql.NullString{}, database.NullString(""), `"" is stored as NULL`)
	assert.Equal(t, sql.NullString{String: "a", Valid: true}, database.NullString("a"))
}
