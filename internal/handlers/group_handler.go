package handlers

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/noteduco342/OMMessenger-backend/internal/service"
	"github.com/noteduco342/OMMessenger-backend/internal/validation"
)

type GroupHandler struct {
	groupService *service.GroupService
}

func NewGroupHandler(groupService *service.GroupService) *GroupHandler {
	return &GroupHandler{groupService: groupService}
}

type CreateGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
	Handle      string `json:"handle"`
}

func (h *GroupHandler) CreateGroup(c *fiber.Ctx) error {
	var req CreateGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Group name is required"})
	}

	if req.IsPublic {
		req.Handle = validation.NormalizeHandle(req.Handle)
		if !validation.ValidateHandle(req.Handle) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid handle"})
		}
	}

	userID := c.Locals("userID").(uint)
	group, err := h.groupService.CreateGroupWithVisibility(req.Name, req.Description, userID, req.IsPublic, req.Handle)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "handle") || strings.Contains(msg, "public") || strings.Contains(msg, "taken") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create group"})
	}

	return c.Status(fiber.StatusCreated).JSON(group)
}

func (h *GroupHandler) GetMyGroups(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	groups, err := h.groupService.GetUserGroups(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch groups"})
	}

	return c.JSON(groups)
}

func (h *GroupHandler) JoinGroup(c *fiber.Ctx) error {
	groupID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid group ID"})
	}

	userID := c.Locals("userID").(uint)
	if err := h.groupService.JoinGroup(uint(groupID), userID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Joined group successfully"})
}

func (h *GroupHandler) LeaveGroup(c *fiber.Ctx) error {
	groupID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid group ID"})
	}

	userID := c.Locals("userID").(uint)
	if err := h.groupService.LeaveGroup(uint(groupID), userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to leave group"})
	}

	return c.JSON(fiber.Map{"message": "Left group successfully"})
}

func (h *GroupHandler) GetGroupMembers(c *fiber.Ctx) error {
	groupID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid group ID"})
	}

	group, err := h.groupService.GetGroup(uint(groupID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Group not found"})
	}

	userID := c.Locals("userID").(uint)
	if !group.IsPublic {
		isMember, err := h.groupService.IsMember(uint(groupID), userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to check membership"})
		}
		if !isMember {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
		}
	}

	members, err := h.groupService.GetGroupMembers(uint(groupID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch members"})
	}

	return c.JSON(members)
}

func (h *GroupHandler) SearchPublicGroups(c *fiber.Ctx) error {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Search query is required"})
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		l := c.QueryInt("limit", 20)
		if l > 0 && l <= 50 {
			limit = l
		}
	}

	groups, err := h.groupService.SearchPublicGroups(query, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to search groups"})
	}

	return c.JSON(fiber.Map{"groups": groups})
}

func (h *GroupHandler) GetPublicGroupByHandle(c *fiber.Ctx) error {
	handle := validation.NormalizeHandle(c.Params("handle"))
	if handle == "" || !validation.ValidateHandle(handle) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid handle"})
	}

	group, err := h.groupService.GetPublicGroupByHandle(handle)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Group not found"})
	}

	return c.JSON(group)
}

func (h *GroupHandler) JoinPublicGroupByHandle(c *fiber.Ctx) error {
	handle := validation.NormalizeHandle(c.Params("handle"))
	if handle == "" || !validation.ValidateHandle(handle) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid handle"})
	}

	userID := c.Locals("userID").(uint)
	group, err := h.groupService.JoinGroupByHandle(handle, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(group)
}

type CreateInviteLinkRequest struct {
	SingleUse        bool `json:"single_use"`
	ExpiresInSeconds *int `json:"expires_in_seconds"`
}

func (h *GroupHandler) CreateInviteLink(c *fiber.Ctx) error {
	groupID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid group ID"})
	}

	var req CreateInviteLinkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var expiresAt *time.Time
	if req.ExpiresInSeconds != nil && *req.ExpiresInSeconds > 0 {
		t := time.Now().Add(time.Duration(*req.ExpiresInSeconds) * time.Second)
		expiresAt = &t
	}

	userID := c.Locals("userID").(uint)
	link, err := h.groupService.CreateInviteLink(uint(groupID), userID, req.SingleUse, expiresAt)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	joinPath := "/join/" + link.Token
	joinURL := ""
	if base := strings.TrimRight(strings.TrimSpace(os.Getenv("PUBLIC_JOIN_BASE_URL")), "/"); base != "" {
		joinURL = base + joinPath
	}

	return c.JSON(fiber.Map{
		"token":      link.Token,
		"join_path":  joinPath,
		"join_url":   joinURL,
		"expires_at": link.ExpiresAt,
		"max_uses":   link.MaxUses,
	})
}

func (h *GroupHandler) JoinByInviteLink(c *fiber.Ctx) error {
	token := strings.TrimSpace(c.Params("token"))
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid token"})
	}
	userID := c.Locals("userID").(uint)
	group, err := h.groupService.JoinGroupByInvite(token, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(group)
}

func (h *GroupHandler) GetInvitePreview(c *fiber.Ctx) error {
	token := strings.TrimSpace(c.Params("token"))
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid token"})
	}

	link, group, err := h.groupService.GetInvitePreview(token)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Minimal preview response (no members, no creator details)
	return c.JSON(fiber.Map{
		"group": fiber.Map{
			"id":          group.ID,
			"name":        group.Name,
			"description": group.Description,
			"icon":        group.Icon,
			"is_public":   group.IsPublic,
			"handle":      group.Handle,
		},
		"expires_at":    link.ExpiresAt,
		"max_uses":      link.MaxUses,
		"used_count":    link.UsedCount,
		"requires_auth": true,
	})
}
