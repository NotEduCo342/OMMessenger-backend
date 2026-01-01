package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/noteduco342/OMMessenger-backend/internal/httpx"
	"github.com/noteduco342/OMMessenger-backend/internal/service"
	"github.com/noteduco342/OMMessenger-backend/internal/validation"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var input service.RegisterInput
	if err := c.BodyParser(&input); err != nil {
		return httpx.BadRequest(c, "invalid_request_body", "Invalid request body")
	}

	input.Email = validation.NormalizeEmail(input.Email)
	input.Username = validation.NormalizeUsername(input.Username)
	input.FullName = validation.TrimAndLimit(input.FullName, 80)

	if !validation.ValidateEmail(input.Email) {
		return httpx.BadRequest(c, "invalid_email", "Invalid email")
	}
	if !validation.ValidateUsername(input.Username) {
		return httpx.BadRequest(c, "invalid_username", "Invalid username")
	}
	if !validation.ValidatePassword(input.Password) {
		return httpx.BadRequest(c, "weak_password", "Password is too short")
	}

	session, err := h.authService.Register(input)
	if err != nil {
		return httpx.BadRequest(c, "registration_failed", err.Error())
	}

	setAuthCookies(c, session.AccessToken, session.RefreshToken)
	log.Printf("event=register user_id=%d email=%s ip=%s rid=%s", session.User.ID, session.User.Email, c.IP(), c.Locals("requestid"))

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user": session.User,
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var input service.LoginInput
	if err := c.BodyParser(&input); err != nil {
		return httpx.BadRequest(c, "invalid_request_body", "Invalid request body")
	}
	input.Email = validation.NormalizeEmail(input.Email)
	if input.Email == "" || input.Password == "" {
		return httpx.BadRequest(c, "missing_fields", "Email and password are required")
	}

	session, err := h.authService.Login(input)
	if err != nil {
		log.Printf("event=login_failed email=%s ip=%s rid=%s", input.Email, c.IP(), c.Locals("requestid"))
		return httpx.Unauthorized(c, "invalid_credentials", "Invalid credentials")
	}

	setAuthCookies(c, session.AccessToken, session.RefreshToken)
	log.Printf("event=login user_id=%d email=%s ip=%s rid=%s", session.User.ID, session.User.Email, c.IP(), c.Locals("requestid"))

	return c.JSON(fiber.Map{
		"user": session.User,
	})
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	refreshToken := c.Cookies("om_refresh")
	session, err := h.authService.RefreshSession(refreshToken)
	if err != nil {
		log.Printf("event=refresh_failed ip=%s rid=%s", c.IP(), c.Locals("requestid"))
		return httpx.Unauthorized(c, "invalid_refresh", "Invalid refresh token")
	}

	setAuthCookies(c, session.AccessToken, session.RefreshToken)
	log.Printf("event=refresh user_id=%d ip=%s rid=%s", session.User.ID, c.IP(), c.Locals("requestid"))
	return c.JSON(fiber.Map{
		"user": session.User,
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	refreshToken := c.Cookies("om_refresh")
	_ = h.authService.Logout(refreshToken)

	clearCookie(c, "om_access")
	clearCookie(c, "om_refresh")

	return c.JSON(fiber.Map{
		"ok": true,
	})
}

// CSRF issues/rotates a CSRF token cookie for browser clients.
func (h *AuthHandler) CSRF(c *fiber.Ctx) error {
	setCSRFCookie(c)
	return c.JSON(fiber.Map{"ok": true})
}

func setAuthCookies(c *fiber.Ctx, accessToken string, refreshToken string) {
	cookieSecure := os.Getenv("COOKIE_SECURE") == "true"
	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	sameSite := os.Getenv("COOKIE_SAMESITE")
	if sameSite == "" {
		sameSite = "Lax"
	}

	access := &fiber.Cookie{
		Name:     "om_access",
		Value:    accessToken,
		HTTPOnly: true,
		Secure:   cookieSecure,
		SameSite: sameSite,
		Path:     "/",
		Expires:  time.Now().Add(15 * time.Minute),
	}
	if cookieDomain != "" {
		access.Domain = cookieDomain
	}
	c.Cookie(access)

	refresh := &fiber.Cookie{
		Name:     "om_refresh",
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   cookieSecure,
		SameSite: sameSite,
		Path:     "/",
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	}
	if cookieDomain != "" {
		refresh.Domain = cookieDomain
	}
	c.Cookie(refresh)

	setCSRFCookie(c)
}

func setCSRFCookie(c *fiber.Ctx) {
	cookieSecure := os.Getenv("COOKIE_SECURE") == "true"
	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	sameSite := os.Getenv("COOKIE_SAMESITE")
	if sameSite == "" {
		sameSite = "Lax"
	}

	// 32 bytes => 43 chars base64url
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return
	}
	token := base64.RawURLEncoding.EncodeToString(b)
	// keep cookie small and simple
	token = strings.TrimSpace(token)

	csrf := &fiber.Cookie{
		Name:     "om_csrf",
		Value:    token,
		HTTPOnly: false,
		Secure:   cookieSecure,
		SameSite: sameSite,
		Path:     "/",
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	}
	if cookieDomain != "" {
		csrf.Domain = cookieDomain
	}
	c.Cookie(csrf)
}

func clearCookie(c *fiber.Ctx, name string) {
	cookieSecure := os.Getenv("COOKIE_SECURE") == "true"
	cookieDomain := os.Getenv("COOKIE_DOMAIN")

	dead := &fiber.Cookie{
		Name:     name,
		Value:    "",
		HTTPOnly: true,
		Secure:   cookieSecure,
		SameSite: "Lax",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	}
	if cookieDomain != "" {
		dead.Domain = cookieDomain
	}
	c.Cookie(dead)
}

func (h *AuthHandler) GetCurrentUser(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	// In a real app, fetch user from database
	return c.JSON(fiber.Map{
		"user_id": userID,
	})
}
