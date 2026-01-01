package middleware

import (
	"crypto/subtle"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/noteduco342/OMMessenger-backend/internal/httpx"
)

// CSRFRequired protects cookie-authenticated browser requests.
// Modes via CSRF_MODE:
// - token: require X-OM-CSRF header to match om_csrf cookie (default)
// - origin: only enforce Origin allow-list
// - off: disable checks
func CSRFRequired() fiber.Handler {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("CSRF_MODE")))
	if mode == "" {
		mode = "token"
	}
	allowedOrigins := splitCSV(strings.TrimSpace(os.Getenv("ALLOWED_ORIGINS")))

	return func(c *fiber.Ctx) error {
		if mode == "off" {
			return c.Next()
		}

		switch c.Method() {
		case fiber.MethodGet, fiber.MethodHead, fiber.MethodOptions:
			return c.Next()
		}

		origin := strings.TrimSpace(c.Get("Origin"))
		if origin == "" {
			// Non-browser clients typically have no Origin; allow.
			return c.Next()
		}

		if len(allowedOrigins) > 0 && !originAllowed(origin, allowedOrigins) {
			return httpx.Forbidden(c, "forbidden_origin", "Origin not allowed")
		}

		if mode == "origin" {
			return c.Next()
		}

		csrfCookie := c.Cookies("om_csrf")
		csrfHeader := c.Get("X-OM-CSRF")
		if csrfCookie == "" || csrfHeader == "" {
			return httpx.Forbidden(c, "csrf_required", "Missing CSRF token")
		}

		if subtle.ConstantTimeCompare([]byte(csrfCookie), []byte(csrfHeader)) != 1 {
			return httpx.Forbidden(c, "csrf_invalid", "Invalid CSRF token")
		}

		return c.Next()
	}
}
