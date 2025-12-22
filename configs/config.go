package configs

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	App 		AppConfig
	Database 	DatabaseConfig
	JWT 		JWTConfig
	OTP 		OTPConfig
}

type AppConfig struct {
	Name 	string
	Env 	string
	Port 	string
}

type DatabaseConfig struct {
	Host 		string
	Port 		string
	User 		string
	Password 	string
	Name 		string
	SSLMode 	string
}

type JWTConfig struct {
	Secret 				string
	AccessTokenExpire 	string
	RefreshTokenExpire 	string
}

type OTPConfig struct {
	Expire string
	Length int
}

var NewConfigs *Config

func LoadConfig() error {

	// Set config file
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		log.Println(".env file not found, using environment variables")
	}

	SetDefaults()

	config := &Config{
		App: AppConfig{
            Name: viper.GetString("APP_NAME"),
            Env:  viper.GetString("APP_ENV"),
            Port: viper.GetString("APP_PORT"),
        },
        Database: DatabaseConfig{
            Host:     viper.GetString("DB_HOST"),
            Port:     viper.GetString("DB_PORT"),
            User:     viper.GetString("DB_USER"),
            Password: viper.GetString("DB_PASSWORD"),
            Name:     viper.GetString("DB_NAME"),
            SSLMode:  viper.GetString("DB_SSLMODE"),
        },
        JWT: JWTConfig{
            Secret:             viper.GetString("JWT_SECRET"),
            AccessTokenExpire:  viper.GetString("JWT_ACCESS_TOKEN_EXPIRE"),
            RefreshTokenExpire: viper.GetString("JWT_REFRESH_TOKEN_EXPIRE"),
        },
        OTP: OTPConfig{
            Expire: viper.GetString("OTP_EXPIRE"),
            Length: viper.GetInt("OTP_LENGTH"),
        },
	}

 	// Validate required fields
    if err := validateConfig(config); err != nil {
        return err
    }

    NewConfigs = config
    log.Println("✅ Configuration loaded successfully")
    return nil
}

func SetDefaults() {
 	viper.SetDefault("APP_NAME", "auth-system")
    viper.SetDefault("APP_ENV", "development")
    viper.SetDefault("APP_PORT", "8080")

    viper.SetDefault("DB_HOST", "localhost")
    viper.SetDefault("DB_PORT", "5432")
    viper.SetDefault("DB_SSLMODE", "disable")

    viper.SetDefault("JWT_ACCESS_TOKEN_EXPIRE", "15m")
    viper.SetDefault("JWT_REFRESH_TOKEN_EXPIRE", "7d")

    viper.SetDefault("OTP_EXPIRE", "5m")
    viper.SetDefault("OTP_LENGTH", 5)
}

// validateConfig to validate mandatory configs
func validateConfig(config *Config) error {
    if config.Database.User == "" {
        return fmt.Errorf("DB_USER is required")
    }
    if config.Database.Password == "" {
        return fmt.Errorf("DB_PASSWORD is required")
    }
    if config.Database.Name == "" {
        return fmt.Errorf("DB_NAME is required")
    }
    if config.JWT.Secret == "" {
        return fmt.Errorf("JWT_SECRET is required")
    }

    return nil
}

// GetConfig mengembalikan config yang sudah di-load
func GetConfig() *Config {
    return NewConfigs
}
