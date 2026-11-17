//go:build integration

package pgtest_test

import (
	"testing"

	"github.com/letenk/golang-authentication/internal/testutil/pgtest"
	"github.com/stretchr/testify/assert"
)

func TestOpen_AppliesMigrationsAndTruncates(t *testing.T) {
	db := pgtest.Open(t)

	id := pgtest.InsertUser(t, db.DB, "a@example.com", "")

	assert.Equal(t, int64(1), id)
	assert.Equal(t, 1, pgtest.CountRows(t, db.DB, "users"))
}

func TestOpen_SecondCallStartsClean(t *testing.T) {
	db := pgtest.Open(t)

	assert.Equal(t, 0, pgtest.CountRows(t, db.DB, "users"))
	assert.Equal(t, int64(1), pgtest.InsertUser(t, db.DB, "", "+628111"))
}
