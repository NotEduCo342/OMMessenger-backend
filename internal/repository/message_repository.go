package repository

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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

func (r *MessageRepository) Update(message *models.Message) error {
	return r.db.Save(message).Error
}

func (r *MessageRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete any pending messages associated with this message
		if err := tx.Unscoped().Where("message_id = ?", id).Delete(&models.PendingMessage{}).Error; err != nil {
			return err
		}
		// Then delete the message
		return tx.Unscoped().Delete(&models.Message{}, id).Error
	})
}

func (r *MessageRepository) FindByID(id uint) (*models.Message, error) {
	var message models.Message
	err := r.db.Preload("Sender").Preload("ReplyTo").First(&message, id).Error
	return &message, err
}

func (r *MessageRepository) FindConversation(userID1, userID2 uint, limit int) ([]models.Message, error) {
	var messages []models.Message
	var clearedAt *time.Time
	var dc models.DeletedConversation
	convoID := fmt.Sprintf("user_%d", userID2)
	if err := r.db.Where("user_id = ? AND conversation_id = ?", userID1, convoID).First(&dc).Error; err == nil {
		clearedAt = &dc.ClearedAt
	}

	query := r.db.Preload("Sender").Preload("ReplyTo").
		Where("(sender_id = ? AND recipient_id = ?) OR (sender_id = ? AND recipient_id = ?)",
			userID1, userID2, userID2, userID1).
		Where("NOT (recipient_id = ? AND recipient_blocked = ?)", userID1, true)

	if clearedAt != nil {
		query = query.Where("created_at > ?", *clearedAt)
	}

	err := query.Order("id DESC").Limit(limit).Find(&messages).Error

	return messages, err
}

// FindConversationCursor fetches messages using cursor-based pagination (more efficient)
func (r *MessageRepository) FindConversationCursor(userID1, userID2 uint, cursor uint, limit int) ([]models.Message, error) {
	var messages []models.Message
	var clearedAt *time.Time
	var dc models.DeletedConversation
	convoID := fmt.Sprintf("user_%d", userID2)
	if err := r.db.Where("user_id = ? AND conversation_id = ?", userID1, convoID).First(&dc).Error; err == nil {
		clearedAt = &dc.ClearedAt
	}

	query := r.db.Preload("Sender").Preload("ReplyTo").
		Where("(sender_id = ? AND recipient_id = ?) OR (sender_id = ? AND recipient_id = ?)",
			userID1, userID2, userID2, userID1).
		Where("NOT (recipient_id = ? AND recipient_blocked = ?)", userID1, true)

	if cursor > 0 {
		query = query.Where("id < ?", cursor)
	}
	if clearedAt != nil {
		query = query.Where("created_at > ?", *clearedAt)
	}

	err := query.Order("id DESC").Limit(limit).Find(&messages).Error

	return messages, err
}

// FindGroupMessages fetches group messages with cursor-based pagination
func (r *MessageRepository) FindGroupMessages(requestingUserID uint, groupID uint, cursor uint, limit int) ([]models.Message, error) {
	var messages []models.Message
	var clearedAt *time.Time
	var dc models.DeletedConversation
	convoID := fmt.Sprintf("group_%d", groupID)
	if err := r.db.Where("user_id = ? AND conversation_id = ?", requestingUserID, convoID).First(&dc).Error; err == nil {
		clearedAt = &dc.ClearedAt
	}

	query := r.db.Preload("Sender").Preload("ReplyTo").Where("group_id = ?", groupID)

	if cursor > 0 {
		query = query.Where("id < ?", cursor)
	}
	if clearedAt != nil {
		query = query.Where("created_at > ?", *clearedAt)
	}

	err := query.Order("id DESC").Limit(limit).Find(&messages).Error
	return messages, err
}

// GetLatestDirectMessageID returns the latest message ID in a DM conversation (0 if none)
func (r *MessageRepository) GetLatestDirectMessageID(userID1, userID2 uint) (uint, error) {
	var maxID uint
	err := r.db.Model(&models.Message{}).
		Where("group_id IS NULL").
		Where("(sender_id = ? AND recipient_id = ?) OR (sender_id = ? AND recipient_id = ?)",
			userID1, userID2, userID2, userID1).
		Select("COALESCE(MAX(id), 0)").
		Scan(&maxID).Error
	return maxID, err
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

func (r *MessageRepository) MarkConversationAsRead(userID uint, peerID uint) (int64, error) {
	tx := r.db.Model(&models.Message{}).
		Where("group_id IS NULL").
		Where("recipient_id = ?", userID).
		Where("sender_id = ?", peerID).
		Where("is_read = false").
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": gorm.Expr("NOW()"),
			"status":  models.StatusRead,
		})
	return tx.RowsAffected, tx.Error
}

// FindByClientID finds a message by client ID and sender
func (r *MessageRepository) FindByClientID(clientID string, senderID uint) (*models.Message, error) {
	var message models.Message
	err := r.db.Preload("Sender").Preload("ReplyTo").
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

	// Look up clear timestamp
	var clearedAt *time.Time
	var dc models.DeletedConversation
	if err := r.db.Where("user_id = ? AND conversation_id = ?", requestingUserID, conversationID).First(&dc).Error; err == nil {
		clearedAt = &dc.ClearedAt
	}

	query := r.db.Preload("Sender").Preload("ReplyTo").Where("messages.id > ?", lastMessageID)
	if clearedAt != nil {
		query = query.Where("messages.created_at > ?", *clearedAt)
	}

	switch kind {
	case "user":
		otherUserID := id
		query = query.
			Where("messages.group_id IS NULL").
			Where("(messages.sender_id = ? AND messages.recipient_id = ?) OR (messages.sender_id = ? AND messages.recipient_id = ?)",
				requestingUserID, otherUserID, otherUserID, requestingUserID).
			Where("NOT (messages.recipient_id = ? AND messages.recipient_blocked = ?)", requestingUserID, true)
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

func (r *MessageRepository) DeleteConversationForEveryone(userID1, userID2 uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Get all message IDs in this DM
		var msgIDs []uint
		if err := tx.Model(&models.Message{}).
			Where("group_id IS NULL").
			Where("(sender_id = ? AND recipient_id = ?) OR (sender_id = ? AND recipient_id = ?)",
				userID1, userID2, userID2, userID1).
			Pluck("id", &msgIDs).Error; err != nil {
			return err
		}

		if len(msgIDs) > 0 {
			// Delete pending messages associated with these messages
			if err := tx.Unscoped().Where("message_id IN ?", msgIDs).Delete(&models.PendingMessage{}).Error; err != nil {
				return err
			}
		}

		// Hard-delete messages from both sides
		if err := tx.Unscoped().
			Where("group_id IS NULL").
			Where("(sender_id = ? AND recipient_id = ?) OR (sender_id = ? AND recipient_id = ?)",
				userID1, userID2, userID2, userID1).
			Delete(&models.Message{}).Error; err != nil {
			return err
		}

		// Also clean up any DeletedConversation records for both sides
		convoID1 := fmt.Sprintf("user_%d", userID2)
		convoID2 := fmt.Sprintf("user_%d", userID1)
		if err := tx.Where("(user_id = ? AND conversation_id = ?) OR (user_id = ? AND conversation_id = ?)",
			userID1, convoID1, userID2, convoID2).
			Delete(&models.DeletedConversation{}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *MessageRepository) ClearConversationForUser(userID uint, conversationID string) error {
	var dc models.DeletedConversation
	err := r.db.Where("user_id = ? AND conversation_id = ?", userID, conversationID).First(&dc).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			dc = models.DeletedConversation{
				UserID:         userID,
				ConversationID: conversationID,
				ClearedAt:      time.Now(),
			}
			return r.db.Create(&dc).Error
		}
		return err
	}
	dc.ClearedAt = time.Now()
	return r.db.Save(&dc).Error
}

// GetLatestGroupMessageID returns the latest message ID in a group (0 if none)
func (r *MessageRepository) GetLatestGroupMessageID(groupID uint) (uint, error) {
	var maxID uint
	err := r.db.Model(&models.Message{}).
		Where("group_id = ?", groupID).
		Select("COALESCE(MAX(id), 0)").
		Scan(&maxID).Error
	return maxID, err
}

// IsMessageInGroup checks whether a message belongs to a group
func (r *MessageRepository) IsMessageInGroup(messageID uint, groupID uint) (bool, error) {
	var dummy int
	// Optimization: Use Select("1").Limit(1) instead of Count() to fast-fail on exist checks
	err := r.db.Model(&models.Message{}).
		Select("1").
		Where("id = ? AND group_id = ?", messageID, groupID).
		Limit(1).
		Scan(&dummy).Error
	return dummy == 1, err
}

func (r *MessageRepository) IsBlocked(userID1, userID2 uint) (bool, error) {
	var dummy int
	// Optimization: Use Select("1").Limit(1) instead of Count() to fast-fail on exist checks
	err := r.db.Table("blocks").
		Select("1").
		Where("(blocker_id = ? AND blocked_id = ?) OR (blocker_id = ? AND blocked_id = ?)", userID1, userID2, userID2, userID1).
		Limit(1).
		Scan(&dummy).Error
	return dummy == 1, err
}

func (r *MessageRepository) IsBlockedBy(blockerID, blockedID uint) (bool, error) {
	var dummy int
	// Optimization: Use Select("1").Limit(1) instead of Count() to fast-fail on exist checks
	err := r.db.Table("blocks").
		Select("1").
		Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).
		Limit(1).
		Scan(&dummy).Error
	return dummy == 1, err
}

