//go:build integration

package database_test

import (
	"context"
	"testing"

	"github.com/letenk/golang-authentication/configs/database"
	"github.com/letenk/golang-authentication/internal/testutil/pgtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckStandardConformingStrings_On(t *testing.T) {
	db := pgtest.Open(t)

	assert.NoError(t, database.CheckStandardConformingStrings(context.Background(), db.DB))
}

// Bun interpolates values into the SQL text and only doubles single quotes,
// which is injection-safe only while standard_conforming_strings is on.
func TestCheckStandardConformingStrings_OffIsRefused(t *testing.T) {
	db := pgtest.Open(t)
	ctx := context.Background()

	conn, err := db.DB.Conn(ctx)
	require.NoError(t, err)
	defer conn.Close()
	_, err = conn.ExecContext(ctx, "SET standard_conforming_strings = off")
	require.NoError(t, err)
	defer func() { _, _ = conn.ExecContext(ctx, "RESET standard_conforming_strings") }()

	err = database.CheckStandardConformingStrings(ctx, conn)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "standard_conforming_strings")
}
