package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/noteduco342/OMMessenger-backend/internal/httpx"
	"github.com/noteduco342/OMMessenger-backend/internal/service"
)

type VersionHandler struct {
	versionService *service.VersionService
}

func NewVersionHandler(versionService *service.VersionService) *VersionHandler {
	return &VersionHandler{
		versionService: versionService,
	}
}

// GetVersion returns the latest app version for a platform
// Public endpoint - no authentication required
// GET /api/version?platform=android
func (h *VersionHandler) GetVersion(c *fiber.Ctx) error {
	// Get platform from query (default: android)
	platform := c.Query("platform", "android")

	// Optional: Log version check for analytics
	userAgent := c.Get("User-Agent", "unknown")
	log.Printf("Version check: platform=%s, user_agent=%s, ip=%s",
		platform, userAgent, c.IP())

	// Get latest version from service
	version, err := h.versionService.GetLatestVersion(platform)
	if err != nil {
		// Return neutral response if version not found (don't block app)
		log.Printf("Version check failed for platform %s: %v", platform, err)

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"version":             "1.0.0",
			"build_number":        1,
			"download_url":        "",
			"changelog":           "",
			"force_update":        false,
			"min_supported_build": 1,
		})
	}

	// Return version information
	return c.JSON(version.ToResponse())
}

// CheckUpdate determines if client needs to update
// Optional endpoint for more detailed update checking
// GET /api/version/check?platform=android&build=1
func (h *VersionHandler) CheckUpdate(c *fiber.Ctx) error {
	platform := c.Query("platform", "android")
	currentBuild := c.QueryInt("build", 0)

	if currentBuild == 0 {
		return httpx.BadRequest(c, "missing_build", "build parameter is required")
	}

	needsUpdate, latestVersion, err := h.versionService.CheckUpdateRequired(platform, currentBuild)
	if err != nil {
		log.Printf("Update check failed: %v", err)
		return httpx.Internal(c, "update_check_failed")
	}

	isForceUpdate, err := h.versionService.IsForceUpdateRequired(platform, currentBuild)
	if err != nil {
		log.Printf("Force update check failed: %v", err)
		isForceUpdate = false // Fail safe - don't force update on error
	}

	return c.JSON(fiber.Map{
		"needs_update":   needsUpdate,
		"force_update":   isForceUpdate,
		"current_build":  currentBuild,
		"latest_version": latestVersion.ToResponse(),
	})
}
