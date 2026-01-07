package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/noteduco342/OMMessenger-backend/internal/cache"
	"github.com/noteduco342/OMMessenger-backend/internal/httpx"
	"github.com/noteduco342/OMMessenger-backend/internal/models"
	"github.com/noteduco342/OMMessenger-backend/internal/service"
	"github.com/noteduco342/OMMessenger-backend/internal/validation"
)

type MessageHandler struct {
	messageService *service.MessageService
	messageCache   *cache.MessageCache
}

func NewMessageHandler(messageService *service.MessageService, messageCache *cache.MessageCache) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
		messageCache:   messageCache,
	}
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
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Check for cursor-based pagination
	var messages []models.Message
	if cursorStr := c.Query("cursor"); cursorStr != "" {
		cursor, err := strconv.ParseUint(cursorStr, 10, 32)
		if err != nil {
			return httpx.BadRequest(c, "invalid_cursor", "Invalid cursor")
		}
		messages, err = h.messageService.GetConversationCursor(userID, uint(recipientID), uint(cursor), limit)
		if err != nil {
			return httpx.Internal(c, "fetch_messages_failed")
		}
	} else {
		// Try cache first (only for non-cursor requests)
		if cached, ok := h.messageCache.GetConversation(userID, uint(recipientID)); ok && len(cached) > 0 {
			messages = cached
			// Limit cached results
			if len(messages) > limit {
				messages = messages[:limit]
			}
		} else {
			messages, err = h.messageService.GetConversation(userID, uint(recipientID), limit)
			if err != nil {
				return httpx.Internal(c, "fetch_messages_failed")
			}
			// Cache the result
			if len(messages) > 0 {
				_ = h.messageCache.SetConversation(userID, uint(recipientID), messages)
			}
		}
	}

	// Convert to response format
	responses := make([]interface{}, len(messages))
	for i, msg := range messages {
		responses[i] = msg.ToResponse()
	}

	// Add cursor info for pagination
	result := fiber.Map{
		"messages": responses,
		"count":    len(messages),
	}

	if len(messages) > 0 {
		// Messages are returned newest-first.
		// Use the last element (oldest in this page) as the cursor for loading older messages.
		result["next_cursor"] = messages[len(messages)-1].ID
	}

	return c.JSON(result)
}

func (h *MessageHandler) GetConversations(c *fiber.Ctx) error {
	userID, err := httpx.LocalUint(c, "userID")
	if err != nil {
		return httpx.Unauthorized(c, "unauthorized", "Unauthorized")
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	var cursorCreatedAt *time.Time
	var cursorMessageID uint
	if cursorCreatedAtStr := c.Query("cursor_created_at"); cursorCreatedAtStr != "" {
		parsed, err := time.Parse(time.RFC3339Nano, cursorCreatedAtStr)
		if err != nil {
			parsed2, err2 := time.Parse(time.RFC3339, cursorCreatedAtStr)
			if err2 != nil {
				return httpx.BadRequest(c, "invalid_cursor_created_at", "Invalid cursor_created_at")
			}
			parsed = parsed2
		}
		cursorCreatedAt = &parsed

		cursorMessageID64, err := strconv.ParseUint(c.Query("cursor_message_id"), 10, 32)
		if err != nil || cursorMessageID64 == 0 {
			return httpx.BadRequest(c, "invalid_cursor_message_id", "Invalid cursor_message_id")
		}
		cursorMessageID = uint(cursorMessageID64)
	}

	rows, err := h.messageService.ListDirectConversations(userID, cursorCreatedAt, cursorMessageID, limit)
	if err != nil {
		return httpx.Internal(c, "fetch_conversations_failed")
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	conversations := make([]interface{}, 0, len(rows))
	for _, r := range rows {
		conversations = append(conversations, fiber.Map{
			"peer": fiber.Map{
				"id":        r.PeerID,
				"username":  r.PeerUsername,
				"email":     r.PeerEmail,
				"full_name": r.PeerFullName,
				"avatar":    r.PeerAvatar,
				"is_online": r.PeerIsOnline,
				"last_seen": r.PeerLastSeen,
			},
			"unread_count":  r.UnreadCount,
			"last_activity": r.LastActivity,
			"last_message": fiber.Map{
				"id":        r.MessageID,
				"client_id": r.MessageClientID,
				"sender_id": r.MessageSenderID,
				"sender": fiber.Map{
					"id":        r.SenderID,
					"username":  r.SenderUsername,
					"email":     r.SenderEmail,
					"full_name": r.SenderFullName,
					"avatar":    r.SenderAvatar,
					"is_online": r.SenderIsOnline,
					"last_seen": r.SenderLastSeen,
				},
				"recipient_id": r.MessageRecipientID,
				"group_id":     nil,
				"content":      r.MessageContent,
				"message_type": r.MessageType,
				"status":       r.MessageStatus,
				"is_delivered": r.MessageIsDelivered,
				"is_read":      r.MessageIsRead,
				"created_at":   r.MessageCreatedAt,
			},
		})
	}

	result := fiber.Map{
		"conversations": conversations,
		"count":         len(conversations),
	}
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		result["next_cursor_created_at"] = last.MessageCreatedAt.Format(time.RFC3339Nano)
		result["next_cursor_message_id"] = last.MessageID
	}

	return c.JSON(result)
}
