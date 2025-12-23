package credential

import (
	"flag"

	"github.com/fsnotify/fsnotify"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/spf13/viper"
)

const (
	pathCredentialNameDefault = "."
	fileCredentialName        = ".env" // ← Explicit
	fileCredentialType        = "env"
)

func InitCredentialEnv(e *echo.Echo) {

	var credentialConfigPath string
	flag.StringVar(&credentialConfigPath, "credentials-path", pathCredentialNameDefault, "your credential credentials path config, default /credential")

	flag.Parse()

	credential := GetCredential()
	credential.SetConfigName(fileCredentialName)
	credential.AddConfigPath(credentialConfigPath)
	credential.SetConfigType(fileCredentialType)

	log.Debugf("credential file : " + credential.ConfigFileUsed())
	err := credential.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Warn("No .env file found, using defaults and environment variables")
		} else {
			e.Logger.Fatal("Error reading config file:", err)
			panic(err)
		}
	}

 	// Validate required configs
    if err := ValidateRequiredConfig(); err != nil {
        e.Logger.Fatal("Config validation failed:", err)
        panic(err)
    }

    log.Info("Config validation passed!")

	initDefaultCredential()

	credential.WatchConfig()
	log.Infof("initialized WatchConfig(): success : credential")
	credential.OnConfigChange(func(e fsnotify.Event) {
		log.Infof("Config file changed:", e.Name)
	})

	log.Infof("initialized configs viper: success : credential")
}

func initDefaultCredential() {
	credential := GetCredential()

	credential.SetDefault("APP_NAME", "auth-system")
	credential.SetDefault("APP_ENV", "development")
	credential.SetDefault("APP_PORT", "8080")

	credential.SetDefault("DB_HOST", "localhost")
	credential.SetDefault("DB_PORT", "5432")
	credential.SetDefault("DB_SSLMODE", "disable")

	credential.SetDefault("JWT_ACCESS_TOKEN_EXPIRE", "15m")
	credential.SetDefault("JWT_REFRESH_TOKEN_EXPIRE", "7d")

	credential.SetDefault("OTP_EXPIRE", "5m")
	credential.SetDefault("OTP_LENGTH", 5)
}
