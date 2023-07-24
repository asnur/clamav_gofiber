package main

import (
	"github.com/asnur/clamav_gofiber/command"
	"github.com/gofiber/fiber/v2"
)

func main() {
	c := command.NewClamd("tcp://localhost:3310")

	if err := c.Ping(); err != nil {
		panic(err)
	}

	app := fiber.New()

	app.Post("/scan", func(ctx *fiber.Ctx) error {
		files, err := ctx.FormFile("file")
		if err != nil {
			return err
		}

		file, err := files.Open()

		if err != nil {
			return err
		}

		var abort chan bool
		ch, err := c.ScanStream(file, abort)
		if err != nil {
			return err
		}

		response := <-ch

		return ctx.JSON(response)
	})

	if err := app.Listen(":3000"); err != nil {
		panic(err)
	}
}
