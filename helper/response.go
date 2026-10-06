package helper

import (
	"github.com/gofiber/fiber/v2"
	"uts-044/model"
)

func Success(c *fiber.Ctx, status int, message string, data any) error {
	return c.Status(status).JSON(model.Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func SuccessWithMeta(c *fiber.Ctx, status int, message string, data any, meta any) error {
	return c.Status(status).JSON(model.Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

func Fail(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(model.Response{
		Success: false,
		Message: message,
	})
}

func FailValidation(c *fiber.Ctx, errors any) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(model.Response{
		Success: false,
		Message: "Validasi gagal",
		Errors:  errors,
	})
}

func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}
