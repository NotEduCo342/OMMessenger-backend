package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/noteduco342/OMMessenger-backend/internal/service"
)

type MessageHandler struct {
	messageService *service.MessageService
}

func NewMessageHandler(messageService *service.MessageService) *MessageHandler {
	return &MessageHandler{messageService: messageService}
}

func (h *MessageHandler) SendMessage(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	
	var input service.SendMessageInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if input.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Content is required",
		})
	}

	message, err := h.messageService.SendMessage(userID, input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to send message",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(message.ToResponse())
}

func (h *MessageHandler) GetMessages(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	
	recipientIDStr := c.Query("recipient_id")
	if recipientIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "recipient_id is required",
		})
	}

	recipientID, err := strconv.ParseUint(recipientIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid recipient_id",
		})
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	messages, err := h.messageService.GetConversation(userID, uint(recipientID), limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch messages",
		})
	}

	// Convert to response format
	responses := make([]interface{}, len(messages))
	for i, msg := range messages {
		responses[i] = msg.ToResponse()
	}

	return c.JSON(fiber.Map{
		"messages": responses,
	})
}
