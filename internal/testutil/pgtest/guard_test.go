package pgtest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequireTestDatabase(t *testing.T) {
	assert.NoError(t, requireTestDatabase("auth_test"))
	assert.NoError(t, requireTestDatabase("authentication_db_test"))

	for _, name := range []string{"authentication_db", "auth_testing", "test", "postgres", ""} {
		assert.Error(t, requireTestDatabase(name), "%q must be refused: Open truncates tables", name)
	}
}
