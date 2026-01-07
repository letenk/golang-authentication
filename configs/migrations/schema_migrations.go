package migrations

import (
	"fmt"

	"github.com/letenk/golang-authentication/configs/credential"
)

func MigrationConnection() string {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		credential.GetString("db.configs.host"),
		credential.GetString("db.configs.port"),
		credential.GetString("db.configs.user"),
		credential.GetString("db.configs.password"),
		credential.GetString("db.configs.name"),
		credential.GetString("db.configs.sslmode"),
	)

	return dsn
}

func MigrationPath() string {
	return "./migrations"
}

// ValidateConnection checks if DB config is valid
func ValidateConnection() error {
	required := map[string]string{
		"db.configs.host":     credential.GetString("db.configs.host"),
		"db.configs.port":     credential.GetString("db.configs.port"),
		"db.configs.user":     credential.GetString("db.configs.user"),
		"db.configs.password": credential.GetString("db.configs.password"),
		"db.configs.name":     credential.GetString("db.configs.name"),
	}

	for key, value := range required {
		if value == "" {
			return fmt.Errorf("required config %s is not set", key)
		}
	}

	return nil
}
