package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/noteduco342/OMMessenger-backend/internal/httpx"
	"github.com/noteduco342/OMMessenger-backend/internal/service"
	"github.com/noteduco342/OMMessenger-backend/internal/validation"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// CheckUsername checks if a username is available
func (h *UserHandler) CheckUsername(c *fiber.Ctx) error {
	username := c.Query("username")
	if username == "" {
		return httpx.BadRequest(c, "missing_username", "Username is required")
	}
	username = validation.NormalizeUsername(username)
	if !validation.ValidateUsername(username) {
		return httpx.BadRequest(c, "invalid_username", "Invalid username")
	}

	available, err := h.userService.IsUsernameAvailable(username)
	if err != nil {
		return httpx.Internal(c, "check_username_failed")
	}

	return c.JSON(fiber.Map{
		"available": available,
	})
}

// UpdateProfile updates user profile information
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	userID, err := httpx.LocalUint(c, "userID")
	if err != nil {
		return httpx.Unauthorized(c, "unauthorized", "Unauthorized")
	}

	var input service.UpdateProfileInput
	if err := c.BodyParser(&input); err != nil {
		return httpx.BadRequest(c, "invalid_request_body", "Invalid request body")
	}
	if input.Username != "" {
		u := validation.NormalizeUsername(input.Username)
		if !validation.ValidateUsername(u) {
			return httpx.BadRequest(c, "invalid_username", "Invalid username")
		}
		input.Username = u
	}
	if input.FullName != "" {
		input.FullName = validation.TrimAndLimit(input.FullName, 80)
	}

	user, err := h.userService.UpdateProfile(userID, input)
	if err != nil {
		return httpx.BadRequest(c, "update_profile_failed", err.Error())
	}

	return c.JSON(fiber.Map{
		"user": user.ToResponse(),
	})
}

// GetCurrentUser gets the authenticated user's profile
func (h *UserHandler) GetCurrentUser(c *fiber.Ctx) error {
	userID, err := httpx.LocalUint(c, "userID")
	if err != nil {
		return httpx.Unauthorized(c, "unauthorized", "Unauthorized")
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		return httpx.Unauthorized(c, "unauthorized", "Unauthorized")
	}

	return c.JSON(fiber.Map{
		"user": user.ToResponse(),
	})
}
