package credential

import (
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

type credential struct {
	viperCredential *viper.Viper
}

var cred *credential

func init() {
	cred = new(credential)
	cred.viperCredential = viper.New()
}

func (c *credential) getCredential(key string) any {
	return c.viperCredential.Get(key)
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
	return cred.viperCredential
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
