package credential

import (
	"fmt"

	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

type credential struct {
	vCredential *viper.Viper
}

var cred *credential

func init() {
	cred = new(credential)
	cred.vCredential = viper.New()
}

func (c *credential) getCredential(key string) any {
	return c.vCredential.Get(key)
}

func getOrFallback(key string) any {
	output := cred.getCredential(key)

	isEmpty := output == nil || output == "" || output == " "
	if isEmpty {
		return viper.GetViper().Get(key)
	}

	return output
}

func GetCredential() *viper.Viper {
	return cred.vCredential
}

func GetString(key string) string {
	output := getOrFallback(key)
	return cast.ToString(output)
}

func GetBool(key string) bool {
	output := getOrFallback(key)
	return cast.ToBool(output)
}

func GetInt(key string) int {
	output := getOrFallback(key)
	return cast.ToInt(output)
}

func GetStringWithDefault(key, defaultValue string) string {
    value := GetString(key)
    if value == "" {
        return defaultValue
    }
    return value
}

func GetIntWithDefault(key string, defaultValue int) int {
    value := GetInt(key)
    if value == 0 {
        return defaultValue
    }
    return value
}

func ValidateRequiredConfig() error {
    required := []string{
        "DB_USER",
        "DB_PASSWORD", 
        "DB_NAME",
        "JWT_SECRET",
    }
    
    for _, key := range required {
        if GetString(key) == "" {
            return fmt.Errorf("required config %s is not set", key)
        }
    }
    
    return nil
}
