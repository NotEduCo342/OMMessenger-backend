package httpx

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type ErrorResponse struct {
	Error     string `json:"error"`
	Code      string `json:"code,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

func requestID(c *fiber.Ctx) string {
	if v := c.Locals("requestid"); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func Error(c *fiber.Ctx, status int, code string, message string) error {
	if message == "" {
		message = "Request failed"
	}
	return c.Status(status).JSON(ErrorResponse{
		Error:     message,
		Code:      code,
		RequestID: requestID(c),
	})
}

func BadRequest(c *fiber.Ctx, code string, message string) error {
	return Error(c, fiber.StatusBadRequest, code, message)
}

func Unauthorized(c *fiber.Ctx, code string, message string) error {
	return Error(c, fiber.StatusUnauthorized, code, message)
}

func Forbidden(c *fiber.Ctx, code string, message string) error {
	return Error(c, fiber.StatusForbidden, code, message)
}

func Internal(c *fiber.Ctx, code string) error {
	return Error(c, fiber.StatusInternalServerError, code, "Internal server error")
}

func LocalUint(c *fiber.Ctx, key string) (uint, error) {
	v := c.Locals(key)
	if v == nil {
		return 0, fmt.Errorf("missing local %s", key)
	}
	u, ok := v.(uint)
	if !ok {
		return 0, fmt.Errorf("invalid local %s", key)
	}
	return u, nil
}
