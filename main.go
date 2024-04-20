package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func main() {
	app := fiber.New()

	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Hello, World")
	})

	err := app.Listen(":3000")
	if err != nil {
		log.Error(err)
		return
	}
}
