package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/noteduco342/OMMessenger-backend/internal/service"
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
}

func (h *GroupHandler) CreateGroup(c *fiber.Ctx) error {
	var req CreateGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Group name is required"})
	}

	userID := c.Locals("userID").(uint)
	group, err := h.groupService.CreateGroup(req.Name, req.Description, userID)
	if err != nil {
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

	members, err := h.groupService.GetGroupMembers(uint(groupID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch members"})
	}

	return c.JSON(members)
}
