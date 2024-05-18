package credential

import (
	"github.com/fsnotify/fsnotify"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

const (
	fileCredentialType = ".env"
)

func InitCredentialEnv(f *fiber.App) {

	credential := GetCredential()
	credential.AddConfigPath(".")
	credential.SetConfigFile(fileCredentialType)

	log.Debugf("credential file : " + credential.ConfigFileUsed())
	err := credential.ReadInConfig()
	if err != nil {
		log.Fatal(err)
		panic(f)
	}

	credential.WatchConfig()
	log.Infof("initialized WatchConfig(): success : credential")
	credential.OnConfigChange(func(e fsnotify.Event) {
		log.Infof("Config file changed:", e.Name)
	})

	log.Infof("initialized configs viper: success", "/."+fileCredentialType)
}
