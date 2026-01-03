package migrations

import (
	"fmt"

	"github.com/letenk/golang-authentication/configs/credential"
)

func MigrationConnection() string {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		credential.GetString("DB_HOST"),
		credential.GetString("DB_PORT"),
		credential.GetString("DB_USER"),
		credential.GetString("DB_PASSWORD"),
		credential.GetString("DB_NAME"),
		credential.GetString("DB_SSLMODE"),
	)

	return dsn
}

func MigrationPath() string {
	return "./migrations"
}


// ValidateConnection checks if DB config is valid
func ValidateConnection() error {
    required := map[string]string{
        "DB_HOST":     credential.GetString("DB_HOST"),
        "DB_PORT":     credential.GetString("DB_PORT"),
        "DB_USER":     credential.GetString("DB_USER"),
        "DB_PASSWORD": credential.GetString("DB_PASSWORD"),
        "DB_NAME":     credential.GetString("DB_NAME"),
    }
    
    for key, value := range required {
        if value == "" {
            return fmt.Errorf("required config %s is not set", key)
        }
    }
    
    return nil
}
