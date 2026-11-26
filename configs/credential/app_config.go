package credential

import (
	"fmt"
	"slices"
	"strings"
)

type AppConfig struct {
	Application ApplicationConfig `mapstructure:"application"`
	Database    DatabaseConfig    `mapstructure:"db"`
	Auth        AuthConfig        `mapstructure:"auth"`
	Email       EmailConfig       `mapstructure:"email"`
}

type ApplicationConfig struct {
	Name string `mapstructure:"name"`
	Mode string `mapstructure:"mode"`
	Env  string `mapstructure:"env"`
	Port string `mapstructure:"port"`
	// Comma-separated origins allowed to call the API from a browser, with cookies.
	CORSAllowedOrigins string `mapstructure:"cors_allowed_origins"`
}

// AllowedOrigins returns CORSAllowedOrigins split on commas, trimmed, without empty entries.
func (a ApplicationConfig) AllowedOrigins() []string {
	var origins []string
	for _, origin := range strings.Split(a.CORSAllowedOrigins, ",") {
		if origin = strings.TrimSpace(origin); origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}

type DatabaseConfig struct {
	Configs DBConnectionConfig `mapstructure:"configs"`
}

type DBConnectionConfig struct {
	Host        string `mapstructure:"host"`
	Port        string `mapstructure:"port"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	Name        string `mapstructure:"name"`
	SSLMode     string `mapstructure:"sslmode"`
	MaxIdleConn int    `mapstructure:"maxIdleConn"`
	MaxOpenConn int    `mapstructure:"maxOpenConn"`
}

type AuthConfig struct {
	JWT JWTConfig `mapstructure:"jwt"`
	OTP OTPConfig `mapstructure:"otp"`
}

type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}

type EmailConfig struct {
	SMTP SMTPConfig `mapstructure:"smtp"`
}

type JWTConfig struct {
	Secret             string `mapstructure:"secret"`
	AccessTokenExpire  string `mapstructure:"access_token_expire"`
	RefreshTokenExpire string `mapstructure:"refresh_token_expire"`
}

type OTPConfig struct {
	Expire string `mapstructure:"expire"`
	Length int    `mapstructure:"length"`
}

func (db *DBConnectionConfig) GetDSNPostgreSQL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&TimeZone=UTC",
		db.User,
		db.Password,
		db.Host,
		db.Port,
		db.Name,
		db.SSLMode,
	)
}

func NewAppConfig() *AppConfig {
	return &AppConfig{
		Application: ApplicationConfig{
			Name: "auth-system",
			Mode: "dev",
			Port: "8080",
		},
		Database: DatabaseConfig{
			Configs: DBConnectionConfig{
				Host:    "localhost",
				Port:    "5432",
				SSLMode: "disable",
			},
		},
		Auth: AuthConfig{
			JWT: JWTConfig{
				AccessTokenExpire:  "15m",
				RefreshTokenExpire: "7d",
			},
			OTP: OTPConfig{
				Expire: "5m",
				Length: 5,
			},
		},
	}
}

// validEnvs lists the accepted values of application.env. Matching is exact on
// purpose: a typo such as "prod" must stop the app, not run it as non-production.
var validEnvs = []string{"development", "staging", "production"}

func (c *AppConfig) Validate() error {
	if !slices.Contains(validEnvs, c.Application.Env) {
		return fmt.Errorf("application.env must be one of %s, got %q",
			strings.Join(validEnvs, ", "), c.Application.Env)
	}

	// Echo treats an empty list as "*" and "*" inside an origin as a wildcard,
	// so production must name every origin exactly.
	if c.Application.Env == "production" {
		origins := c.Application.AllowedOrigins()
		if len(origins) == 0 || slices.ContainsFunc(origins, func(o string) bool { return strings.Contains(o, "*") }) {
			return fmt.Errorf("application.cors_allowed_origins must list exact origins without wildcards in production, got %q",
				c.Application.CORSAllowedOrigins)
		}
	}

	if c.Database.Configs.User == "" {
		return fmt.Errorf("db.configs.user is required")
	}

	if c.Database.Configs.Password == "" {
		return fmt.Errorf("db.configs.password is required")
	}

	if c.Database.Configs.Name == "" {
		return fmt.Errorf("db.configs.name is required")
	}

	return nil
}
