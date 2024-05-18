package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/letenk/golang-authentication/configs/credential"
)

func main() {
	app := fiber.New()
	credential.InitCredentialEnv(app)

	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Hello, World")
	})

	port := credential.GetString("application.port")

	err := app.Listen(":" + port)
	if err != nil {
		log.Error(err)
		return
	}

	fmt.Println("App started 🔥")
}
