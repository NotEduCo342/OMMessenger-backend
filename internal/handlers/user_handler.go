package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/noteduco342/OMMessenger-backend/internal/httpx"
	"github.com/noteduco342/OMMessenger-backend/internal/models"
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
		"user": h.toUserResponse(c, user),
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

	// ETag allows clients to re-check frequently without re-downloading.
	etag := fmt.Sprintf("W/\"u-%d-%d\"", user.ID, user.UpdatedAt.UTC().UnixNano())
	c.Set("ETag", etag)
	c.Set("Cache-Control", "private, max-age=0, must-revalidate")

	if inm := strings.TrimSpace(c.Get("If-None-Match")); inm != "" {
		// Support quoted, weak, and multi-value headers.
		inmNorm := strings.Trim(strings.TrimPrefix(inm, "W/"), "\"")
		etagNorm := strings.Trim(strings.TrimPrefix(etag, "W/"), "\"")
		if strings.Contains(inmNorm, etagNorm) {
			return c.SendStatus(fiber.StatusNotModified)
		}
	}

	return c.JSON(fiber.Map{
		"user": h.toUserResponse(c, user),
	})
}

// SearchUsers searches for users by username or full name
func (h *UserHandler) SearchUsers(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return httpx.BadRequest(c, "missing_query", "Search query is required")
	}

	if len(query) < 3 {
		return c.JSON(fiber.Map{
			"users": []interface{}{},
		})
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		l := c.QueryInt("limit", 10)
		if l > 0 && l <= 10 {
			limit = l
		}
	}

	users, err := h.userService.SearchUsers(query, limit)
	if err != nil {
		return httpx.Internal(c, "search_users_failed")
	}

	// Convert to response format
	responses := make([]interface{}, len(users))
	for i := range users {
		responses[i] = h.toUserResponse(c, &users[i])
	}

	return c.JSON(fiber.Map{
		"users": responses,
	})
}

// GetUserByUsername gets a user's public profile by username
func (h *UserHandler) GetUserByUsername(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		return httpx.BadRequest(c, "missing_username", "Username is required")
	}

	user, err := h.userService.GetUserByUsername(username)
	if err != nil {
		return httpx.BadRequest(c, "user_not_found", "User not found")
	}

	return c.JSON(fiber.Map{
		"user": h.toUserResponse(c, user),
	})
}

// GetUser gets a user's profile by ID or username.
// Route: GET /users/:identifier
// - If identifier is numeric => lookup by user ID
// - Otherwise => lookup by username
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	identifier := strings.TrimSpace(c.Params("identifier"))
	if identifier == "" {
		return httpx.BadRequest(c, "missing_identifier", "Identifier is required")
	}

	// Numeric path segment => treat as user ID
	if id64, err := strconv.ParseUint(identifier, 10, 64); err == nil {
		if id64 == 0 {
			return httpx.BadRequest(c, "invalid_user_id", "Invalid user ID")
		}
		user, err := h.userService.GetUserByID(uint(id64))
		if err != nil {
			return httpx.Error(c, fiber.StatusNotFound, "user_not_found", "User not found")
		}
		return c.JSON(fiber.Map{
			"user": h.toUserResponse(c, user),
		})
	}

	// Otherwise treat as username
	user, err := h.userService.GetUserByUsername(identifier)
	if err != nil {
		return httpx.Error(c, fiber.StatusNotFound, "user_not_found", "User not found")
	}

	return c.JSON(fiber.Map{
		"user": h.toUserResponse(c, user),
	})
}

func (h *UserHandler) toUserResponse(c *fiber.Ctx, user *models.User) models.UserResponse {
	resp := user.ToResponse()
	userID, err := httpx.LocalUint(c, "userID")
	if err == nil {
		isBlocked, _ := h.userService.IsBlocker(userID, user.ID)
		resp.IsBlocked = isBlocked

		isBlockedByPeer, _ := h.userService.IsBlocker(user.ID, userID)
		if isBlockedByPeer {
			resp.Avatar = ""
			resp.IsOnline = false
			resp.LastSeen = nil
		}
	}
	return resp
}

// BlockUser blocks a user
func (h *UserHandler) BlockUser(c *fiber.Ctx) error {
	userID, err := httpx.LocalUint(c, "userID")
	if err != nil {
		return httpx.Unauthorized(c, "unauthorized", "Unauthorized")
	}

	targetID64, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil || targetID64 == 0 {
		return httpx.BadRequest(c, "invalid_user_id", "Invalid user ID")
	}
	targetID := uint(targetID64)

	if userID == targetID {
		return httpx.BadRequest(c, "cannot_block_self", "You cannot block yourself")
	}

	err = h.userService.BlockUser(userID, targetID)
	if err != nil {
		return httpx.BadRequest(c, "block_failed", err.Error())
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "User blocked successfully",
	})
}

// UnblockUser unblocks a user
func (h *UserHandler) UnblockUser(c *fiber.Ctx) error {
	userID, err := httpx.LocalUint(c, "userID")
	if err != nil {
		return httpx.Unauthorized(c, "unauthorized", "Unauthorized")
	}

	targetID64, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil || targetID64 == 0 {
		return httpx.BadRequest(c, "invalid_user_id", "Invalid user ID")
	}
	targetID := uint(targetID64)

	err = h.userService.UnblockUser(userID, targetID)
	if err != nil {
		return httpx.BadRequest(c, "unblock_failed", err.Error())
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "User unblocked successfully",
	})
}

// GetBlockList returns users blocked by the current user
func (h *UserHandler) GetBlockList(c *fiber.Ctx) error {
	userID, err := httpx.LocalUint(c, "userID")
	if err != nil {
		return httpx.Unauthorized(c, "unauthorized", "Unauthorized")
	}

	users, err := h.userService.GetBlockedUsers(userID)
	if err != nil {
		return httpx.Internal(c, "get_blocks_failed")
	}

	responses := make([]models.UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
		responses[i].IsBlocked = true
	}

	return c.JSON(fiber.Map{
		"users": responses,
		"count": len(responses),
	})
}
