package credential

import (
	"flag"
	"fmt"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/labstack/gommon/log"
	"github.com/spf13/viper"
)

const (
	pathCredentialNameDefault = "."
	fileCredentialName        = ".env" // ← Explicit
	fileCredentialType        = "env"
)

func InitCredentialEnv() error {

	var credentialConfigPath string
	flag.StringVar(&credentialConfigPath, "credentials-path", pathCredentialNameDefault, "credential config path")
	flag.Parse()

	credential := GetCredential()

	credential.AutomaticEnv()
	credential.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	credential.SetConfigName(fileCredentialName)
	credential.AddConfigPath(credentialConfigPath)
	credential.SetConfigType(fileCredentialType)

	log.Debugf("credential file :" + credential.ConfigFileUsed())
	err := credential.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Warn("No .env file found, using defaults and environment variables")
			initDefaultCredential()
		} else {
			return fmt.Errorf("read config failed: %w", err)
		}
	}

	// Validate required configs
	if err := ValidateRequiredConfig(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	log.Info("Config validation passed!")

	Config = &AppConfig{}
	 if err := credential.Unmarshal(Config); err != nil {
        return fmt.Errorf("failed to unmarshal config: %w", err)
    }

	credential.WatchConfig()
	log.Infof("initialized WatchConfig(): success : credential")
	credential.OnConfigChange(func(e fsnotify.Event) {
		log.Infof("Config file changed: %s", e.Name)

		// Re-unmarshal on config change
		if err := credential.Unmarshal(Config); err != nil {
            log.Errorf("Failed to reload config: %v", err)
        } else {
            log.Info("Config reloaded successfully!")
        }
	})

	log.Infof("initialized configs viper: success : credential")

	return nil
}

func initDefaultCredential() {
	credential := GetCredential()

	credential.SetDefault("application.name", "auth-system")
	credential.SetDefault("application.mode", "dev")
	credential.SetDefault("application.port", "8080")

	credential.SetDefault("db.configs.host", "localhost")
	credential.SetDefault("db.configs.port", "5432")
	credential.SetDefault("db.configs.sslmode", "disable")

	credential.SetDefault("auth.jwt.access_token_expire", "15m")
	credential.SetDefault("auth.jwt.refresh_token_expire", "7d")

	credential.SetDefault("auth.otp.expire", "5m")
	credential.SetDefault("auth.otp.length", 5)
}
