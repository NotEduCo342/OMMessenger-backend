package handlers

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/noteduco342/OMMessenger-backend/internal/cache"
	"github.com/noteduco342/OMMessenger-backend/internal/handlers/ws"
	"github.com/noteduco342/OMMessenger-backend/internal/httpx"
	"github.com/noteduco342/OMMessenger-backend/internal/models"
	"github.com/noteduco342/OMMessenger-backend/internal/service"
	"github.com/noteduco342/OMMessenger-backend/internal/validation"
	"gorm.io/gorm"
)

type MessageHandler struct {
	messageService *service.MessageService
	groupService   *service.GroupService
	messageCache   *cache.MessageCache
	hub            *ws.Hub
}

func NewMessageHandler(messageService *service.MessageService, groupService *service.GroupService, messageCache *cache.MessageCache, hub *ws.Hub) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
		groupService:   groupService,
		messageCache:   messageCache,
		hub:            hub,
	}
}

type SendGroupMessageRequest struct {
	ClientID    string `json:"client_id"`
	Content     string `json:"content"`
	MessageType string `json:"message_type"`
}

type MarkGroupReadRequest struct {
	LastReadMessageID uint `json:"last_read_message_id"`
}

type SyncConversationState struct {
	ConversationID string `json:"conversation_id"`
	LastMessageID  uint   `json:"last_message_id"`
}

type SyncMessagesRequest struct {
	Conversations []SyncConversationState `json:"conversations"`
	Limit         int                     `json:"limit"`
}

func parseMessageType(input string) models.MessageType {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "", string(models.TextMessage):
		return models.TextMessage
	case string(models.ImageMessage):
		return models.ImageMessage
	case string(models.FileMessage):
		return models.FileMessage
	default:
		return models.TextMessage
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

func (h *MessageHandler) GetGroupMessages(c *fiber.Ctx) error {
	userID, err := httpx.LocalUint(c, "userID")
	if err != nil {
		return httpx.Unauthorized(c, "unauthorized", "Unauthorized")
	}

	groupID64, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil || groupID64 == 0 {
		return httpx.BadRequest(c, "invalid_group_id", "Invalid group id")
	}
	groupID := uint(groupID64)

	if h.groupService != nil {
		isMember, err := h.groupService.IsMember(groupID, userID)
		if err != nil {
			return httpx.Internal(c, "check_membership_failed")
		}
		if !isMember {
			return httpx.Forbidden(c, "not_group_member", "Not a group member")
		}
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	var messages []models.Message
	if cursorStr := c.Query("cursor"); cursorStr != "" {
		cursor, err := strconv.ParseUint(cursorStr, 10, 32)
		if err != nil {
			return httpx.BadRequest(c, "invalid_cursor", "Invalid cursor")
		}
		messages, err = h.messageService.GetGroupMessages(groupID, uint(cursor), limit)
		if err != nil {
			return httpx.Internal(c, "fetch_messages_failed")
		}
	} else {
		if cached, ok := h.messageCache.GetGroupConversation(groupID); ok && len(cached) > 0 {
			messages = cached
			if len(messages) > limit {
				messages = messages[:limit]
			}
		} else {
			messages, err = h.messageService.GetGroupMessages(groupID, 0, limit)
			if err != nil {
				return httpx.Internal(c, "fetch_messages_failed")
			}
			if len(messages) > 0 {
				_ = h.messageCache.SetGroupConversation(groupID, messages)
			}
		}
	}

	responses := make([]interface{}, len(messages))
	for i, msg := range messages {
		responses[i] = msg.ToResponse()
	}

	result := fiber.Map{
		"messages": responses,
		"count":    len(messages),
	}
	if len(messages) > 0 {
		result["next_cursor"] = messages[len(messages)-1].ID
	}

	return c.JSON(result)
}

// SyncMessages allows REST-based incremental sync for background polling.
// Body: { "conversations": [{"conversation_id":"user_1","last_message_id":10}], "limit": 100 }
func (h *MessageHandler) SyncMessages(c *fiber.Ctx) error {
	userID, err := httpx.LocalUint(c, "userID")
	if err != nil {
		return httpx.Unauthorized(c, "unauthorized", "Unauthorized")
	}

	var input SyncMessagesRequest
	if err := c.BodyParser(&input); err != nil {
		return httpx.BadRequest(c, "invalid_request_body", "Invalid request body")
	}
	if len(input.Conversations) == 0 {
		return httpx.BadRequest(c, "missing_conversations", "conversations is required")
	}

	limit := input.Limit
	if limit <= 0 || limit > 200 {
		limit = 100
	}

	results := make([]fiber.Map, 0, len(input.Conversations))
	for _, conv := range input.Conversations {
		if strings.TrimSpace(conv.ConversationID) == "" {
			return httpx.BadRequest(c, "invalid_conversation_id", "conversation_id is required")
		}
		messages, err := h.messageService.GetMessagesSince(userID, conv.ConversationID, conv.LastMessageID, limit)
		if err != nil {
			return httpx.Internal(c, "sync_failed")
		}

		responses := make([]models.MessageResponse, len(messages))
		for i := range messages {
			responses[i] = messages[i].ToResponse()
		}

		entry := fiber.Map{
			"conversation_id": conv.ConversationID,
			"messages":        responses,
			"has_more":        len(messages) == limit,
		}
		if len(messages) > 0 {
			entry["next_cursor"] = messages[len(messages)-1].ID
		}
		results = append(results, entry)
	}

	return c.JSON(fiber.Map{
		"results": results,
		"count":   len(results),
	})
}

func (h *MessageHandler) SendGroupMessage(c *fiber.Ctx) error {
	userID, err := httpx.LocalUint(c, "userID")
	if err != nil {
		return httpx.Unauthorized(c, "unauthorized", "Unauthorized")
	}

	groupID64, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil || groupID64 == 0 {
		return httpx.BadRequest(c, "invalid_group_id", "Invalid group id")
	}
	groupID := uint(groupID64)

	if h.groupService != nil {
		isMember, err := h.groupService.IsMember(groupID, userID)
		if err != nil {
			return httpx.Internal(c, "check_membership_failed")
		}
		if !isMember {
			return httpx.Forbidden(c, "not_group_member", "Not a group member")
		}
	}

	var input SendGroupMessageRequest
	if err := c.BodyParser(&input); err != nil {
		return httpx.BadRequest(c, "invalid_request_body", "Invalid request body")
	}

	input.ClientID = strings.TrimSpace(input.ClientID)
	input.Content = validation.TrimAndLimit(input.Content, validation.MaxMessageLength())
	if input.ClientID == "" {
		return httpx.BadRequest(c, "missing_client_id", "client_id is required")
	}
	if input.Content == "" {
		return httpx.BadRequest(c, "missing_content", "Content is required")
	}

	// Idempotent send by client_id
	if existing, err := h.messageService.GetByClientID(input.ClientID, userID); err == nil && existing != nil {
		if existing.GroupID != nil && *existing.GroupID == groupID {
			return c.Status(fiber.StatusCreated).JSON(existing.ToResponse())
		}
		return httpx.BadRequest(c, "client_id_conflict", "client_id already used")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return httpx.Internal(c, "get_message_failed")
	}

	msgType := parseMessageType(input.MessageType)
	message, err := h.messageService.CreateWithClientIDAndType(userID, input.ClientID, nil, &groupID, input.Content, msgType)
	if err != nil {
		return httpx.Internal(c, "send_message_failed")
	}

	// Invalidate group message cache and conversation list cache for members
	if h.messageCache != nil {
		_ = h.messageCache.InvalidateGroupConversation(groupID)
	}

	// Broadcast to group members (also queues for offline users)
	if h.hub != nil && h.groupService != nil {
		members, err := h.groupService.GetGroupMembers(groupID)
		if err == nil {
			for _, member := range members {
				if member.ID == userID {
					continue
				}
				_ = h.hub.SendToUserWithID(member.ID, message.ID, map[string]interface{}{
					"type":    "message",
					"message": message.ToResponse(),
				})
				if h.messageCache != nil {
					_ = h.messageCache.InvalidateConversationList(member.ID)
				}
			}
		}
	}

	if h.messageCache != nil {
		_ = h.messageCache.InvalidateConversationList(userID)
	}

	return c.Status(fiber.StatusCreated).JSON(message.ToResponse())
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

	if cursorCreatedAt == nil && h.messageCache != nil {
		if cached, ok := h.messageCache.GetConversationListPayload(userID); ok {
			return c.JSON(cached)
		}
	}

	fetchLimit := limit + 1
	rows, err := h.messageService.ListConversationsUnified(userID, cursorCreatedAt, cursorMessageID, fetchLimit)
	if err != nil {
		return httpx.Internal(c, "fetch_conversations_failed")
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	conversations := make([]interface{}, 0, len(rows))
	for _, r := range rows {
		var conversationID string
		if r.ConversationType == "group" && r.GroupID.Valid {
			conversationID = "group_" + strconv.FormatInt(r.GroupID.Int64, 10)
		} else if r.PeerID.Valid {
			conversationID = "user_" + strconv.FormatInt(r.PeerID.Int64, 10)
		}

		peer := interface{}(nil)
		if r.PeerID.Valid {
			peer = fiber.Map{
				"id":        uint(r.PeerID.Int64),
				"username":  r.PeerUsername.String,
				"email":     r.PeerEmail.String,
				"full_name": r.PeerFullName.String,
				"avatar":    r.PeerAvatar.String,
				"is_online": r.PeerIsOnline.Bool,
				"last_seen": r.PeerLastSeen,
			}
		}

		group := interface{}(nil)
		if r.GroupID.Valid {
			group = fiber.Map{
				"id":           uint(r.GroupID.Int64),
				"name":         r.GroupName.String,
				"icon":         r.GroupIcon.String,
				"member_count": r.MemberCount.Int64,
			}
		}

		var recipientID interface{} = nil
		if r.MessageRecipientID.Valid {
			recipientID = uint(r.MessageRecipientID.Int64)
		}
		var groupID interface{} = nil
		if r.MessageGroupID.Valid {
			groupID = uint(r.MessageGroupID.Int64)
		}

		conversations = append(conversations, fiber.Map{
			"conversation_id": conversationID,
			"peer":            peer,
			"group":           group,
			"unread_count":    r.UnreadCount,
			"last_activity":   r.LastActivity,
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
				"recipient_id":    recipientID,
				"group_id":        groupID,
				"content":         r.MessageContent,
				"message_type":    r.MessageType,
				"status":          r.MessageStatus,
				"is_delivered":    r.MessageIsDelivered,
				"is_read":         r.MessageIsRead,
				"created_at":      r.MessageCreatedAt,
				"created_at_unix": r.MessageCreatedAt.UTC().Unix(),
			},
		})
	}

	result := fiber.Map{
		"conversations": conversations,
		"count":         len(conversations),
	}
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		result["next_cursor_created_at"] = last.LastActivity.Format(time.RFC3339Nano)
		result["next_cursor_message_id"] = last.MessageID
	}

	if cursorCreatedAt == nil && h.messageCache != nil {
		payload := map[string]interface{}{
			"conversations": conversations,
			"count":         len(conversations),
		}
		if hasMore && len(rows) > 0 {
			payload["next_cursor_created_at"] = result["next_cursor_created_at"]
			payload["next_cursor_message_id"] = result["next_cursor_message_id"]
		}
		_ = h.messageCache.SetConversationListPayload(userID, payload)
	}

	return c.JSON(result)
}

// GetRecentPeers returns recent DM peers for seeding conversation list after reinstall.
// Endpoint: GET /conversations/peers?limit=50
func (h *MessageHandler) GetRecentPeers(c *fiber.Ctx) error {
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

	rows, err := h.messageService.ListRecentPeers(userID, limit)
	if err != nil {
		return httpx.Internal(c, "fetch_recent_peers_failed")
	}

	peers := make([]fiber.Map, 0, len(rows))
	for _, r := range rows {
		peers = append(peers, fiber.Map{
			"peer": fiber.Map{
				"id":        r.PeerID,
				"username":  r.PeerUsername,
				"email":     r.PeerEmail,
				"full_name": r.PeerFullName,
				"avatar":    r.PeerAvatar,
				"is_online": r.PeerIsOnline,
				"last_seen": r.PeerLastSeen,
			},
			"last_message_id": r.MessageID,
			"last_activity":   r.LastActivity,
		})
	}

	return c.JSON(fiber.Map{
		"peers": peers,
		"count": len(peers),
	})
}

func (h *MessageHandler) MarkConversationRead(c *fiber.Ctx) error {
	userID, err := httpx.LocalUint(c, "userID")
	if err != nil {
		return httpx.Unauthorized(c, "unauthorized", "Unauthorized")
	}

	peerStr := c.Params("peer_id")
	peerID64, err := strconv.ParseUint(peerStr, 10, 32)
	if err != nil || peerID64 == 0 {
		return httpx.BadRequest(c, "invalid_peer_id", "Invalid peer_id")
	}

	cleared, err := h.messageService.MarkConversationAsRead(userID, uint(peerID64))
	if err != nil {
		return httpx.Internal(c, "mark_conversation_read_failed")
	}

	lastRead, err := h.messageService.GetLatestDirectMessageID(userID, uint(peerID64))
	if err == nil && lastRead > 0 && h.hub != nil {
		h.hub.SendToUser(uint(peerID64), fiber.Map{
			"type":                 "read_update",
			"conversation_id":      "user_" + strconv.FormatUint(uint64(userID), 10),
			"user_id":              userID,
			"last_read_message_id": lastRead,
		})
		if h.messageCache != nil {
			_ = h.messageCache.InvalidateConversationList(uint(peerID64))
		}
	}
	if h.messageCache != nil {
		_ = h.messageCache.InvalidateConversationList(userID)
	}

	return c.JSON(fiber.Map{
		"ok":      true,
		"cleared": cleared,
	})
}

func (h *MessageHandler) MarkGroupRead(c *fiber.Ctx) error {
	userID, err := httpx.LocalUint(c, "userID")
	if err != nil {
		return httpx.Unauthorized(c, "unauthorized", "Unauthorized")
	}

	groupID64, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil || groupID64 == 0 {
		return httpx.BadRequest(c, "invalid_group_id", "Invalid group id")
	}
	groupID := uint(groupID64)

	if h.groupService != nil {
		isMember, err := h.groupService.IsMember(groupID, userID)
		if err != nil {
			return httpx.Internal(c, "check_membership_failed")
		}
		if !isMember {
			return httpx.Forbidden(c, "not_group_member", "Not a group member")
		}
	}

	var input MarkGroupReadRequest
	if err := c.BodyParser(&input); err != nil {
		return httpx.BadRequest(c, "invalid_request_body", "Invalid request body")
	}

	if input.LastReadMessageID > 0 {
		belongs, err := h.messageService.IsMessageInGroup(input.LastReadMessageID, groupID)
		if err != nil {
			return httpx.Internal(c, "validate_message_failed")
		}
		if !belongs {
			return httpx.BadRequest(c, "invalid_message_id", "Message does not belong to group")
		}
	}

	latestID, err := h.messageService.GetLatestGroupMessageID(groupID)
	if err != nil {
		return httpx.Internal(c, "latest_message_failed")
	}
	lastRead := input.LastReadMessageID
	if lastRead > latestID {
		lastRead = latestID
	}

	if h.groupService != nil {
		if err := h.groupService.UpsertReadStateMonotonic(groupID, userID, lastRead); err != nil {
			return httpx.Internal(c, "mark_group_read_failed")
		}
	}
	if h.messageCache != nil {
		_ = h.messageCache.InvalidateConversationList(userID)
	}

	return c.JSON(fiber.Map{
		"ok":                      true,
		"last_read_message_id":    lastRead,
		"latest_group_message_id": latestID,
	})
}

func (h *MessageHandler) GetGroupReadState(c *fiber.Ctx) error {
	userID, err := httpx.LocalUint(c, "userID")
	if err != nil {
		return httpx.Unauthorized(c, "unauthorized", "Unauthorized")
	}

	groupID64, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil || groupID64 == 0 {
		return httpx.BadRequest(c, "invalid_group_id", "Invalid group id")
	}
	groupID := uint(groupID64)

	if h.groupService != nil {
		isMember, err := h.groupService.IsMember(groupID, userID)
		if err != nil {
			return httpx.Internal(c, "check_membership_failed")
		}
		if !isMember {
			return httpx.Forbidden(c, "not_group_member", "Not a group member")
		}
	}

	myState, err := h.groupService.GetReadState(groupID, userID)
	if err != nil {
		return httpx.Internal(c, "get_read_state_failed")
	}

	states, err := h.groupService.ListReadStates(groupID)
	if err != nil {
		return httpx.Internal(c, "get_read_state_failed")
	}

	members := make([]fiber.Map, 0, len(states))
	for _, s := range states {
		members = append(members, fiber.Map{
			"user_id":              s.UserID,
			"last_read_message_id": s.LastReadMessageID,
		})
	}

	return c.JSON(fiber.Map{
		"my_last_read_message_id": myState.LastReadMessageID,
		"members":                 members,
	})
}
