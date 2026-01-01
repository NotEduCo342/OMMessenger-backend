package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/noteduco342/OMMessenger-backend/internal/httpx"
)

func RequireRole(role string) fiber.Handler {
	role = strings.ToLower(strings.TrimSpace(role))
	return func(c *fiber.Ctx) error {
		v := c.Locals("role")
		userRole, _ := v.(string)
		if strings.ToLower(userRole) != role {
			return httpx.Forbidden(c, "forbidden", "Insufficient permissions")
		}
		return c.Next()
	}
}
