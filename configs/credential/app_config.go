package credential

import "fmt"

type AppConfig struct {
	Application ApplicationConfig `mapstructure:"application"`
	Database    DatabaseConfig    `mapstructure:"db"`
	Auth        AuthConfig        `mapstructure:"auth"`
}

type ApplicationConfig struct {
	Name string `mapstructure:"name"`
	Mode string `mapstructure:"mode"`
	Env  string `mapstructure:"env"`
	Port string `mapstructure:"port"`
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
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
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

func (c *AppConfig) Validate() error {
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

