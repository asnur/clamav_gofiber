package clamav

import (
	"fmt"

	"github.com/asnur/clamav_gofiber/command"
	"github.com/asnur/clamav_gofiber/domain"
	"github.com/gofiber/fiber/v2"
)

type Config struct {
	ClamdAddress string // ClamdAddress is the address of the clamd daemon
	FieldName    string // FieldName is the name of the field in the multipart form
}

func New(config Config) fiber.Handler {

	c := command.NewClamd(config.ClamdAddress)

	if err := c.Ping(); err != nil {
		panic(err)
	}

	return func(ctx *fiber.Ctx) error {
		files, err := ctx.FormFile(config.FieldName)
		if err != nil {
			fmt.Println("No file found")
			return ctx.Next()
		}

		file, err := files.Open()

		if err != nil {
			return ctx.JSON(domain.Response{
				Status:  fiber.StatusBadRequest,
				Message: err.Error(),
				Data:    nil,
			})
		}

		var abort chan bool
		ch, err := c.ScanStream(file, abort)
		if err != nil {
			return ctx.JSON(domain.Response{
				Status:  fiber.StatusBadRequest,
				Message: err.Error(),
				Data:    nil,
			})
		}

		response := <-ch

		if response.Status == domain.RES_FOUND {
			return ctx.Status(fiber.StatusForbidden).JSON(
				domain.Response{
					Status:  fiber.StatusForbidden,
					Message: "File is infected",
					Data:    response,
				},
			)
		}

		return ctx.Next()
	}
}
