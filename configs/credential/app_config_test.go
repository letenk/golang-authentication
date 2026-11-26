package credential

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validConfig returns a config that passes every check, so each test only
// changes the field it is about.
func validConfig(env string) *AppConfig {
	cfg := NewAppConfig()
	cfg.Application.Env = env
	cfg.Database.Configs.User = "postgres"
	cfg.Database.Configs.Password = "secret"
	cfg.Database.Configs.Name = "auth"
	cfg.Application.CORSAllowedOrigins = "https://app.example.com"
	return cfg
}

func TestValidate_AcceptsKnownEnvironments(t *testing.T) {
	for _, env := range []string{"development", "staging", "production"} {
		t.Run(env, func(t *testing.T) {
			assert.NoError(t, validConfig(env).Validate())
		})
	}
}

// Anything that is not exactly one of the known values must stop the app,
// so a typo can never silently run production as a non-production env.
func TestValidate_RejectsUnknownEnvironments(t *testing.T) {
	cases := map[string]string{
		"empty":            "",
		"abbreviation":     "prod",
		"wrong case":       "Production",
		"trailing space":   "production ",
		"unsupported name": "local",
	}

	for name, env := range cases {
		t.Run(name, func(t *testing.T) {
			err := validConfig(env).Validate()

			require.Error(t, err)
			assert.Contains(t, err.Error(), "application.env")
			assert.Contains(t, err.Error(), fmt.Sprintf("%q", env))
		})
	}
}

// Production must list its allowed origins explicitly: a missing value or a
// wildcard would let any site call the API with the user's cookies.
func TestValidate_ProductionRequiresExplicitCORSOrigins(t *testing.T) {
	for name, origins := range map[string]string{
		"empty":            "",
		"only spaces":      " , ",
		"wildcard":         "*",
		"wildcard inside":  "https://app.example.com,*",
		"wildcard pattern": "https://*",
	} {
		t.Run(name, func(t *testing.T) {
			cfg := validConfig("production")
			cfg.Application.CORSAllowedOrigins = origins

			err := cfg.Validate()

			require.Error(t, err)
			assert.Contains(t, err.Error(), "application.cors_allowed_origins")
		})
	}
}

func TestValidate_DevelopmentAllowsEmptyCORSOrigins(t *testing.T) {
	cfg := validConfig("development")
	cfg.Application.CORSAllowedOrigins = ""

	assert.NoError(t, cfg.Validate())
}

func TestAllowedOrigins_SplitsAndTrims(t *testing.T) {
	app := ApplicationConfig{CORSAllowedOrigins: " https://a.example.com , ,https://b.example.com "}

	assert.Equal(t, []string{"https://a.example.com", "https://b.example.com"}, app.AllowedOrigins())
}
