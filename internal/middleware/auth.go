package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/noteduco342/OMMessenger-backend/internal/httpx"
)

type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		var tokenString string
		if authHeader != "" {
			// Extract token from "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return httpx.Unauthorized(c, "invalid_authorization", "Invalid authorization format")
			}
			tokenString = parts[1]
		} else {
			tokenString = c.Cookies("om_access")
		}

		if tokenString == "" {
			return httpx.Unauthorized(c, "missing_access_token", "Missing access token")
		}

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			if token.Method == nil || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			return httpx.Unauthorized(c, "invalid_access_token", "Invalid or expired token")
		}

		// Extract claims
		claims, ok := token.Claims.(*Claims)
		if !ok {
			return httpx.Unauthorized(c, "invalid_access_token", "Invalid token")
		}

		// Store user info in context
		c.Locals("userID", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}
