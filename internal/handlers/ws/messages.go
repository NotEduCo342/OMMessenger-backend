package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/noteduco342/OMMessenger-backend/internal/models"
)

const (
	MsgSync     = "sync"
	MsgChat     = "chat"
	MsgAck      = "ack"
	MsgTyping   = "typing"
	MsgRead     = "read"
	MsgDelivery = "delivery"
)

// ConversationState tracks client's last known message for a conversation
type ConversationState struct {
	ConversationID string `json:"conversation_id"`
	LastMessageID  uint   `json:"last_message_id"`
	LastSeenAt     int64  `json:"last_seen_at"`
}

// MessageSync is sent by client to initiate sync
type MessageSync struct {
	Conversations []ConversationState `json:"conversations"`
}

func (msg *MessageSync) GetType() string {
	return MsgSync
}

func (msg *MessageSync) Process(ctx *MessageContext) error {
	// For each conversation, get messages since last known ID
	for _, conv := range msg.Conversations {
		messages, err := ctx.MessageService.GetMessagesSince(conv.ConversationID, conv.LastMessageID, 100)
		if err != nil {
			log.Printf("Error fetching messages for conversation %s: %v", conv.ConversationID, err)
			continue
		}

		// Send batch response
		response := SyncResponse{
			Type:           "sync_response",
			ConversationID: conv.ConversationID,
			Messages:       messages,
			HasMore:        len(messages) == 100,
		}

		if len(messages) > 0 {
			response.NextCursor = &messages[len(messages)-1].ID
		}

		if err := ctx.Conn.WriteJSON(response); err != nil {
			return err
		}
	}

	return nil
}

type SyncResponse struct {
	Type           string           `json:"type"`
	ConversationID string           `json:"conversation_id"`
	Messages       []models.Message `json:"messages"`
	HasMore        bool             `json:"has_more"`
	NextCursor     *uint            `json:"next_cursor,omitempty"`
}

// MessageChat is a new chat message from client
type MessageChat struct {
	ClientID       string `json:"client_id"` // UUID from client for deduplication
	ConversationID string `json:"conversation_id"`
	RecipientID    *uint  `json:"recipient_id,omitempty"`
	GroupID        *uint  `json:"group_id,omitempty"`
	Content        string `json:"content"`
	MessageType    string `json:"message_type"`
}

func (msg *MessageChat) GetType() string {
	return MsgChat
}

func (msg *MessageChat) Process(ctx *MessageContext) error {
	log.Printf("💬 Processing chat message from user %d: client_id=%s, recipient=%v", ctx.UserID, msg.ClientID, msg.RecipientID)

	if msg.ClientID == "" {
		return SendError(ctx.Conn, "missing_client_id", "client_id is required", "")
	}

	// Check for duplicate using ClientID
	existing, err := ctx.MessageService.GetByClientID(msg.ClientID, ctx.UserID)
	if err == nil && existing != nil {
		// Already processed, send ACK again with proper wrapper
		log.Printf("✅ Duplicate message, re-sending ACK for server_id=%d", existing.ID)
		ackData, _ := json.Marshal(MessageAck{
			ClientID: msg.ClientID,
			ServerID: existing.ID,
			Status:   string(existing.Status),
		})
		ackWrapper := SerializedMessage{
			Type:    MsgAck,
			Payload: json.RawMessage(ackData),
		}
		return ctx.Conn.WriteJSON(ackWrapper)
	}

	// Save message to database
	log.Printf("💾 Saving new message to database...")
	message, err := ctx.MessageService.CreateWithClientID(ctx.UserID, msg.ClientID, msg.RecipientID, msg.GroupID, msg.Content)
	if err != nil {
		log.Printf("❌ Error saving message: %v", err)
		return SendError(ctx.Conn, "save_failed", "Failed to save message", err.Error())
	}
	log.Printf("✅ Message saved with ID=%d", message.ID)

	// Invalidate conversation cache for both sender and recipient
	if msg.RecipientID != nil && ctx.MessageCache != nil {
		_ = ctx.MessageCache.InvalidateConversation(ctx.UserID, *msg.RecipientID)
		_ = ctx.MessageCache.InvalidateConversationList(ctx.UserID)
		_ = ctx.MessageCache.InvalidateConversationList(*msg.RecipientID)
		_ = ctx.MessageCache.InvalidateUnreadCount(*msg.RecipientID, ctx.UserID)
	}

	// Send ACK to sender with proper wrapper
	log.Printf("📤 Sending ACK to sender...")
	ackData, _ := json.Marshal(MessageAck{
		ClientID: msg.ClientID,
		ServerID: message.ID,
		Status:   "sent",
	})
	ackWrapper := SerializedMessage{
		Type:    MsgAck,
		Payload: json.RawMessage(ackData), // Cast to RawMessage
	}
	if err := ctx.Conn.WriteJSON(ackWrapper); err != nil {
		log.Printf("❌ Error sending ACK: %v", err)
		return err
	}
	log.Printf("✅ ACK sent successfully")

	// Forward to recipient if online
	if msg.RecipientID != nil {
		log.Printf("📨 Forwarding message to recipient %d...", *msg.RecipientID)
		ctx.Hub.SendToUserWithID(*msg.RecipientID, message.ID, map[string]interface{}{
			"type":    "message",
			"message": message.ToResponse(),
		})
	} else if msg.GroupID != nil {
		// Broadcast to group members
		members, err := ctx.GroupService.GetGroupMembers(*msg.GroupID)
		if err == nil {
			var memberIDs []uint
			for _, member := range members {
				if member.ID != ctx.UserID {
					memberIDs = append(memberIDs, member.ID)
				}
			}
			ctx.Hub.BroadcastToUsers(memberIDs, map[string]interface{}{
				"type":    "message",
				"message": message.ToResponse(),
			})
		}
	}

	return nil
}

// MessageAck acknowledges message delivery/read status
type MessageAck struct {
	ClientID string `json:"client_id,omitempty"`
	ServerID uint   `json:"server_id,omitempty"`
	Status   string `json:"status"` // sent, delivered, read
}

func (msg *MessageAck) GetType() string {
	return MsgAck
}

func (msg *MessageAck) Process(ctx *MessageContext) error {
	if msg.ServerID == 0 {
		return SendError(ctx.Conn, "missing_server_id", "server_id is required", "")
	}

	switch msg.Status {
	case "delivered":
		return ctx.MessageService.MarkAsDelivered(msg.ServerID)
	case "read":
		return ctx.MessageService.MarkAsRead(msg.ServerID)
	default:
		return SendError(ctx.Conn, "invalid_status", "Invalid status", msg.Status)
	}
}

// MessageTyping indicates user is typing
type MessageTyping struct {
	ConversationID string `json:"conversation_id"`
	RecipientID    *uint  `json:"recipient_id,omitempty"`
	IsTyping       bool   `json:"is_typing"`
}

func (msg *MessageTyping) GetType() string {
	return MsgTyping
}

func (msg *MessageTyping) Process(ctx *MessageContext) error {
	if msg.RecipientID != nil {
		ctx.Hub.SendToUser(*msg.RecipientID, map[string]interface{}{
			"type":      "typing",
			"sender_id": ctx.UserID,
			"is_typing": msg.IsTyping,
			"timestamp": time.Now().Unix(),
		})
	}
	return nil
}

// MessageRead marks message as read
type MessageRead struct {
	MessageID uint `json:"message_id"`
}

func (msg *MessageRead) GetType() string {
	return MsgRead
}

func (msg *MessageRead) Process(ctx *MessageContext) error {
	return ctx.MessageService.MarkAsRead(msg.MessageID)
}

// MessageDelivery marks message as delivered
type MessageDelivery struct {
	MessageID uint `json:"message_id"`
}

func (msg *MessageDelivery) GetType() string {
	return MsgDelivery
}

func (msg *MessageDelivery) Process(ctx *MessageContext) error {
	return ctx.MessageService.MarkAsDelivered(msg.MessageID)
}
