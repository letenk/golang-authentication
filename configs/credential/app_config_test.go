package credential

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validConfig returns a config that passes every other check, so a failure
// can only come from application.env.
func validConfig(env string) *AppConfig {
	cfg := NewAppConfig()
	cfg.Application.Env = env
	cfg.Database.Configs.User = "postgres"
	cfg.Database.Configs.Password = "secret"
	cfg.Database.Configs.Name = "auth"
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
