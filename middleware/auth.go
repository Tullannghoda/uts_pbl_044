package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"uts-044/helper"
)

func Auth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			return helper.Fail(c, fiber.StatusUnauthorized, "Token tidak ditemukan")
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		claims, err := helper.ParseToken(tokenStr)
		if err != nil {
			return helper.Fail(c, fiber.StatusUnauthorized, "Token tidak valid atau kedaluwarsa")
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)
		return c.Next()
	}
}

func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals("role").(string)
		if role != "admin" {
			return helper.Fail(c, fiber.StatusForbidden, "Hanya admin yang diizinkan")
		}
		return c.Next()
	}
}

func MahasiswaOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals("role").(string)
		if role != "mahasiswa" {
			return helper.Fail(c, fiber.StatusForbidden, "Hanya mahasiswa yang diizinkan")
		}
		return c.Next()
	}
}
