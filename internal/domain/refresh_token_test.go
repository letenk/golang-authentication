package domain_test

import (
	"testing"
	"time"

	"github.com/letenk/golang-authentication/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestRefreshToken_IsRevoked(t *testing.T) {
	now := time.Now()

	assert.False(t, (&domain.RefreshToken{}).IsRevoked())
	assert.True(t, (&domain.RefreshToken{RevokedAt: &now}).IsRevoked())
}

func TestRefreshToken_IsExpired(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	assert.True(t, (&domain.RefreshToken{ExpiresAt: now.Add(-time.Second)}).IsExpired(now))
	assert.False(t, (&domain.RefreshToken{ExpiresAt: now.Add(time.Second)}).IsExpired(now))
}
