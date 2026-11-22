package database_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/letenk/golang-authentication/configs/database"
	"github.com/stretchr/testify/assert"
)

func TestTimePtr(t *testing.T) {
	now := time.Now()

	assert.Nil(t, database.TimePtr(sql.NullTime{}))
	if got := database.TimePtr(sql.NullTime{Time: now, Valid: true}); assert.NotNil(t, got) {
		assert.True(t, now.Equal(*got))
	}
}

func TestNullString(t *testing.T) {
	assert.Equal(t, sql.NullString{}, database.NullString(""), `"" is stored as NULL`)
	assert.Equal(t, sql.NullString{String: "a", Valid: true}, database.NullString("a"))
}

func TestNullStringPtr(t *testing.T) {
	s := "a"

	assert.Equal(t, sql.NullString{}, database.NullStringPtr(nil), "nil means no value")
	assert.Equal(t, sql.NullString{String: "a", Valid: true}, database.NullStringPtr(&s))
}
