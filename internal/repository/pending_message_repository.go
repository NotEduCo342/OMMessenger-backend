package repository

import (
	"time"

	"github.com/noteduco342/OMMessenger-backend/internal/models"
	"gorm.io/gorm"
)

type PendingMessageRepository struct {
	db *gorm.DB
}

func NewPendingMessageRepository(db *gorm.DB) *PendingMessageRepository {
	return &PendingMessageRepository{db: db}
}

// Enqueue adds a message to the pending queue for a user
func (r *PendingMessageRepository) Enqueue(userID, messageID uint, payload string, priority int) error {
	pending := &models.PendingMessage{
		UserID:    userID,
		MessageID: messageID,
		Payload:   payload,
		Priority:  priority,
		Attempts:  0,
	}
	return r.db.Create(pending).Error
}

// GetPendingForUser retrieves all pending messages for a user, ordered by priority and creation time
func (r *PendingMessageRepository) GetPendingForUser(userID uint, limit int) ([]models.PendingMessage, error) {
	var pending []models.PendingMessage
	err := r.db.Where("user_id = ?", userID).
		Order("priority DESC, created_at ASC").
		Limit(limit).
		Find(&pending).Error
	return pending, err
}

// GetRetryable gets messages ready for retry (next_retry <= now)
func (r *PendingMessageRepository) GetRetryable(limit int) ([]models.PendingMessage, error) {
	var pending []models.PendingMessage
	now := time.Now()
	err := r.db.Where("next_retry IS NOT NULL AND next_retry <= ?", now).
		Order("priority DESC, next_retry ASC").
		Limit(limit).
		Find(&pending).Error
	return pending, err
}

// MarkAttempted updates the attempt count and next retry time
func (r *PendingMessageRepository) MarkAttempted(id uint, attempts int, nextRetry *time.Time) error {
	now := time.Now()
	updates := map[string]interface{}{
		"attempts":     attempts,
		"last_attempt": now,
		"next_retry":   nextRetry,
	}
	return r.db.Model(&models.PendingMessage{}).Where("id = ?", id).Updates(updates).Error
}

// Delete removes a pending message (after successful delivery)
func (r *PendingMessageRepository) Delete(id uint) error {
	return r.db.Delete(&models.PendingMessage{}, id).Error
}

// DeleteBatch removes multiple pending messages
func (r *PendingMessageRepository) DeleteBatch(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Delete(&models.PendingMessage{}, ids).Error
}

// CountPendingForUser returns the number of pending messages for a user
func (r *PendingMessageRepository) CountPendingForUser(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.PendingMessage{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// CleanupOld removes pending messages older than the specified duration
func (r *PendingMessageRepository) CleanupOld(olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	return r.db.Where("created_at < ?", cutoff).Delete(&models.PendingMessage{}).Error
}
