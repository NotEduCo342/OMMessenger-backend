package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/noteduco342/OMMessenger-backend/internal/httpx"
	"github.com/noteduco342/OMMessenger-backend/internal/service"
	"github.com/noteduco342/OMMessenger-backend/internal/validation"
)

type MessageHandler struct {
	messageService *service.MessageService
}

func NewMessageHandler(messageService *service.MessageService) *MessageHandler {
	return &MessageHandler{messageService: messageService}
}

func (h *MessageHandler) SendMessage(c *fiber.Ctx) error {
	userID, err := httpx.LocalUint(c, "userID")
	if err != nil {
		return httpx.Unauthorized(c, "unauthorized", "Unauthorized")
	}

	var input service.SendMessageInput
	if err := c.BodyParser(&input); err != nil {
		return httpx.BadRequest(c, "invalid_request_body", "Invalid request body")
	}

	input.Content = validation.TrimAndLimit(input.Content, validation.MaxMessageLength())
	if input.Content == "" {
		return httpx.BadRequest(c, "missing_content", "Content is required")
	}
	if input.RecipientID == nil || *input.RecipientID == 0 {
		return httpx.BadRequest(c, "missing_recipient", "recipient_id is required")
	}

	message, err := h.messageService.SendMessage(userID, input)
	if err != nil {
		return httpx.Internal(c, "send_message_failed")
	}

	return c.Status(fiber.StatusCreated).JSON(message.ToResponse())
}

func (h *MessageHandler) GetMessages(c *fiber.Ctx) error {
	userID, err := httpx.LocalUint(c, "userID")
	if err != nil {
		return httpx.Unauthorized(c, "unauthorized", "Unauthorized")
	}

	recipientIDStr := c.Query("recipient_id")
	if recipientIDStr == "" {
		return httpx.BadRequest(c, "missing_recipient", "recipient_id is required")
	}

	recipientID, err := strconv.ParseUint(recipientIDStr, 10, 32)
	if err != nil {
		return httpx.BadRequest(c, "invalid_recipient", "Invalid recipient_id")
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	messages, err := h.messageService.GetConversation(userID, uint(recipientID), limit)
	if err != nil {
		return httpx.Internal(c, "fetch_messages_failed")
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
