package response

import (
	"github.com/gofiber/fiber/v2"
)

type ResponseOKWithDataModel struct {
	Code         int         `json:"code"`
	Data         interface{} `json:"data"`
	ErrorMessage string      `json:"errorMessage"`
}

type ResponseOKModel struct {
	Code         int    `json:"code"`
	ErrorMessage string `json:"errorMessage"`
}

type ResponseErrorModel struct {
	Code         int         `json:"code"`
	ErrorMessage interface{} `json:"errorMessage"`
}

func ResponseOKWithData(c *fiber.Ctx, data interface{}) error {
	response := ResponseOKWithDataModel{
		Code:         200,
		Data:         data,
		ErrorMessage: "",
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func ResponseOK(c *fiber.Ctx, message string) error {
	response := ResponseOKModel{
		Code:         200,
		ErrorMessage: message,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func ResponseCreated(c *fiber.Ctx, data interface{}) error {
	response := ResponseOKWithDataModel{
		Code:         201,
		Data:         data,
		ErrorMessage: "",
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

func ResponseError(c *fiber.Ctx, err interface{}, code int) error {
	response := ResponseErrorModel{
		Code:         code,
		ErrorMessage: err,
	}

	return c.Status(code).JSON(response)
}
