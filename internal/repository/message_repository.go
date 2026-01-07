package repository

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/noteduco342/OMMessenger-backend/internal/models"
	"gorm.io/gorm"
)

type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(message *models.Message) error {
	return r.db.Create(message).Error
}

func (r *MessageRepository) FindByID(id uint) (*models.Message, error) {
	var message models.Message
	err := r.db.Preload("Sender").First(&message, id).Error
	return &message, err
}

func (r *MessageRepository) FindConversation(userID1, userID2 uint, limit int) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.Preload("Sender").
		Where("(sender_id = ? AND recipient_id = ?) OR (sender_id = ? AND recipient_id = ?)",
			userID1, userID2, userID2, userID1).
		Order("id DESC").
		Limit(limit).
		Find(&messages).Error

	return messages, err
}

// FindConversationCursor fetches messages using cursor-based pagination (more efficient)
func (r *MessageRepository) FindConversationCursor(userID1, userID2 uint, cursor uint, limit int) ([]models.Message, error) {
	var messages []models.Message
	query := r.db.Preload("Sender").
		Where("(sender_id = ? AND recipient_id = ?) OR (sender_id = ? AND recipient_id = ?)",
			userID1, userID2, userID2, userID1)

	if cursor > 0 {
		query = query.Where("id < ?", cursor)
	}

	err := query.Order("id DESC").Limit(limit).Find(&messages).Error

	return messages, err
}

func (r *MessageRepository) MarkAsDelivered(messageID uint) error {
	return r.db.Model(&models.Message{}).Where("id = ?", messageID).
		Updates(map[string]interface{}{
			"is_delivered": true,
			"delivered_at": gorm.Expr("NOW()"),
			"status":       models.StatusDelivered,
		}).Error
}

func (r *MessageRepository) MarkAsRead(messageID uint) error {
	return r.db.Model(&models.Message{}).Where("id = ?", messageID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": gorm.Expr("NOW()"),
			"status":  models.StatusRead,
		}).Error
}

// FindByClientID finds a message by client ID and sender
func (r *MessageRepository) FindByClientID(clientID string, senderID uint) (*models.Message, error) {
	var message models.Message
	err := r.db.Preload("Sender").
		Where("client_id = ? AND sender_id = ?", clientID, senderID).
		First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// FindMessagesSince gets messages for a conversation since a specific message ID (optimized with ID index)

func parseConversationID(conversationID string) (kind string, id uint, err error) {
	conversationID = strings.TrimSpace(conversationID)
	if conversationID == "" {
		return "", 0, fmt.Errorf("empty conversation_id")
	}
	if strings.HasPrefix(conversationID, "user_") {
		s := strings.TrimPrefix(conversationID, "user_")
		v, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			return "", 0, fmt.Errorf("invalid user conversation_id: %w", err)
		}
		return "user", uint(v), nil
	}
	if strings.HasPrefix(conversationID, "group_") {
		s := strings.TrimPrefix(conversationID, "group_")
		v, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			return "", 0, fmt.Errorf("invalid group conversation_id: %w", err)
		}
		return "group", uint(v), nil
	}
	return "", 0, fmt.Errorf("unknown conversation_id format")
}

func (r *MessageRepository) FindMessagesSince(requestingUserID uint, conversationID string, lastMessageID uint, limit int) ([]models.Message, error) {
	var messages []models.Message

	if limit <= 0 {
		limit = 100
	}
	if limit > 200 {
		limit = 200
	}

	kind, id, err := parseConversationID(conversationID)
	if err != nil {
		return nil, err
	}

	query := r.db.Preload("Sender").Where("messages.id > ?", lastMessageID)

	switch kind {
	case "user":
		otherUserID := id
		query = query.
			Where("messages.group_id IS NULL").
			Where("(messages.sender_id = ? AND messages.recipient_id = ?) OR (messages.sender_id = ? AND messages.recipient_id = ?)",
				requestingUserID, otherUserID, otherUserID, requestingUserID)
	case "group":
		groupID := id
		// Enforce group membership by joining group_members with requestingUserID.
		query = query.
			Joins("JOIN group_members gm ON gm.group_id = messages.group_id AND gm.user_id = ?", requestingUserID).
			Where("messages.group_id = ?", groupID)
	default:
		return nil, fmt.Errorf("unsupported conversation kind")
	}

	err = query.Order("messages.id ASC").Limit(limit).Find(&messages).Error

	return messages, err
}
