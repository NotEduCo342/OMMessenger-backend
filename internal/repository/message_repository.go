package repository

import (
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
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

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

// FindMessagesSince gets messages for a conversation since a specific message ID
func (r *MessageRepository) FindMessagesSince(conversationID string, lastMessageID uint, limit int) ([]models.Message, error) {
	var messages []models.Message

	// For now, conversationID format is "user1_user2" for DMs or "group_123" for groups
	// Parse and query accordingly - simplified version for direct messages
	err := r.db.Preload("Sender").
		Where("id > ?", lastMessageID).
		Order("id ASC").
		Limit(limit).
		Find(&messages).Error

	return messages, err
}
