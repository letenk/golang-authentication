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
	flag.StringVar(
		&credentialConfigPath, 
		"credentials-path", 
		pathCredentialNameDefault, 
		"credential config path",
	)
	flag.Parse()

	Config = NewAppConfig()

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