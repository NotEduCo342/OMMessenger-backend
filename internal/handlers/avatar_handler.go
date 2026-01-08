package handlers

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/noteduco342/OMMessenger-backend/internal/httpx"
	"github.com/noteduco342/OMMessenger-backend/internal/service"
	"github.com/noteduco342/OMMessenger-backend/internal/storage"
)

type AvatarHandler struct {
	avatarService *service.AvatarService
}

func NewAvatarHandler(avatarService *service.AvatarService) *AvatarHandler {
	return &AvatarHandler{avatarService: avatarService}
}

func publicAPIBaseURL(c *fiber.Ctx) string {
	base := strings.TrimRight(strings.TrimSpace(getenv("PUBLIC_API_BASE_URL")), "/")
	if base != "" {
		return base
	}
	// Fallback: infer from request.
	return strings.TrimRight(c.BaseURL(), "/") + "/api"
}

func (h *AvatarHandler) UploadMyAvatar(c *fiber.Ctx) error {
	userID, err := httpx.LocalUint(c, "userID")
	if err != nil {
		return httpx.Unauthorized(c, "unauthorized", "Unauthorized")
	}

	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		return httpx.BadRequest(c, "missing_avatar", "avatar file is required")
	}

	f, err := fileHeader.Open()
	if err != nil {
		return httpx.BadRequest(c, "invalid_avatar", "Invalid avatar upload")
	}
	defer f.Close()

	user, err := h.avatarService.UploadAvatar(c.Context(), userID, f, publicAPIBaseURL(c))
	if err != nil {
		if errors.Is(err, service.ErrStorageNotConfigured) {
			return httpx.Error(c, fiber.StatusServiceUnavailable, "storage_not_configured", "Storage not configured")
		}
		if errors.Is(err, storage.ErrTooLarge) {
			return httpx.BadRequest(c, "avatar_too_large", "Avatar is too large")
		}
		if errors.Is(err, storage.ErrUnsupported) {
			return httpx.BadRequest(c, "avatar_unsupported", "Unsupported image type")
		}
		if errors.Is(err, storage.ErrInvalidImage) {
			return httpx.BadRequest(c, "avatar_invalid", "Invalid image")
		}
		return httpx.Internal(c, "avatar_upload_failed")
	}

	return c.JSON(fiber.Map{
		"user": user.ToResponse(),
	})
}

func (h *AvatarHandler) DeleteMyAvatar(c *fiber.Ctx) error {
	userID, err := httpx.LocalUint(c, "userID")
	if err != nil {
		return httpx.Unauthorized(c, "unauthorized", "Unauthorized")
	}

	user, err := h.avatarService.DeleteAvatar(c.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrStorageNotConfigured) {
			return httpx.Error(c, fiber.StatusServiceUnavailable, "storage_not_configured", "Storage not configured")
		}
		return httpx.Internal(c, "avatar_delete_failed")
	}

	return c.JSON(fiber.Map{
		"user": user.ToResponse(),
	})
}
