package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/noteduco342/OMMessenger-backend/internal/httpx"
)

func OriginAllowed() fiber.Handler {
	allowedOrigins := splitCSV(strings.TrimSpace(os.Getenv("ALLOWED_ORIGINS")))
	return func(c *fiber.Ctx) error {
		origin := strings.TrimSpace(c.Get("Origin"))
		if origin == "" || len(allowedOrigins) == 0 {
			return c.Next()
		}
		if !originAllowed(origin, allowedOrigins) {
			return httpx.Forbidden(c, "forbidden_origin", "Origin not allowed")
		}
		return c.Next()
	}
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func originAllowed(origin string, allowed []string) bool {
	for _, a := range allowed {
		if a == origin {
			return true
		}
	}
	return false
}
