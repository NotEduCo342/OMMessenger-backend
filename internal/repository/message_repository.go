package repository

import (
	"gorm.io/gorm"
	"github.com/noteduco342/OMMessenger-backend/internal/models"
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
		}).Error
}

func (r *MessageRepository) MarkAsRead(messageID uint) error {
	return r.db.Model(&models.Message{}).Where("id = ?", messageID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": gorm.Expr("NOW()"),
		}).Error
}
