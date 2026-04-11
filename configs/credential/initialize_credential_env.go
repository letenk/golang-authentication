package credential

import (
	"flag"
	"fmt"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/labstack/gommon/log"
	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

const (
	pathCredentialNameDefault = "."
	fileCredentialName        = ".env" // ← Explicit
	fileCredentialType        = "env"
)

func InitCredentialEnv() error {
	var credentialConfigPath string
	flag.StringVar(
		&credentialConfigPath, 
		"credentials-path", 
		pathCredentialNameDefault, 
		"credential config path",
	)
	flag.Parse()

	Config = NewAppConfig()

	if err := gotenv.Load(".env"); err != nil {
		log.Warnf("No .env file found, using OS environment variables only")
	}

	credential := GetCredential()

	// Setup environment variable reading
	credential.AutomaticEnv()
	credential.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	credential.AllowEmptyEnv(true)

	// Bind all environment variables explicitly
	// This ensures Viper reads from environment even when .env file doesn't exist
	bindEnvVariables(credential)

	credential.SetConfigName(fileCredentialName)
	credential.AddConfigPath(credentialConfigPath)
	credential.SetConfigType(fileCredentialType)

	log.Debugf("credential file: %s", credential.ConfigFileUsed())
	err := credential.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Warn("No .env file found, using defaults and environment variables")
		} else {
			return fmt.Errorf("read config failed: %w", err)
		}
	}

	if err := credential.Unmarshal(Config); err != nil {
        return fmt.Errorf("failed to unmarshal config: %w", err)
    }

	// Validate required configs
	if err := Config.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	log.Info("Config validation passed!")

	credential.WatchConfig()
	log.Infof("initialized WatchConfig(): success : credential")
	credential.OnConfigChange(func(e fsnotify.Event) {
		log.Infof("Config file changed: %s", e.Name)

		// Re-unmarshal on config change
		if err := viper.Unmarshal(Config); err != nil {
			log.Errorf("Failed to reload config: %v", err)
			return
		}

		if err := Config.Validate(); err != nil {
			log.Errorf("Config validation failed after reload: %v", err)
			return
		}

		log.Info("Config reloaded successfully!")
	})

	log.Infof("initialized configs viper: success : credential")

	return nil
}

// bindEnvVariables explicitly binds all configuration keys to environment variables
func bindEnvVariables(v *viper.Viper) {
	// Application
	v.BindEnv("application.name")
	v.BindEnv("application.env")
	v.BindEnv("application.port")
	v.BindEnv("application.mode")

	// Database
	v.BindEnv("db.configs.host")
	v.BindEnv("db.configs.port")
	v.BindEnv("db.configs.user")
	v.BindEnv("db.configs.password")
	v.BindEnv("db.configs.name")
	v.BindEnv("db.configs.sslmode")
	v.BindEnv("db.configs.maxIdleConn")
	v.BindEnv("db.configs.maxOpenConn")

	// JWT
	v.BindEnv("auth.jwt.secret")
	v.BindEnv("auth.jwt.access_token_expire")
	v.BindEnv("auth.jwt.refresh_token_expire")

	// OTP
	v.BindEnv("auth.otp.expire")
	v.BindEnv("auth.otp.length")

	// Email SMTP
	v.BindEnv("email.smtp.host")
	v.BindEnv("email.smtp.port")
	v.BindEnv("email.smtp.username")
	v.BindEnv("email.smtp.password")
	v.BindEnv("email.smtp.from")
}