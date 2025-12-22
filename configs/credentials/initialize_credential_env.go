package credentials

import (
	"flag"

	"github.com/labstack/echo"
	"github.com/labstack/echo/v4"
)

const (
	pathCredentialNameDefault 	= "secret"
	fileCredentialType 			= "env"
)

func InitCredentialEnv(e *echo.Echo) {
	
	var credentialConfigPath string
	flag.StringVar(&credentialConfigPath, "credentials-path", pathCredentialNameDefault, "your credential credentials path config, default /credential")

	flag.Parse()
	
	
}