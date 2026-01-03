package models

import (
	"time"

	"gorm.io/gorm"
)

// PendingMessage represents a message queued for delivery to an offline user
type PendingMessage struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Target user who should receive this message
	UserID uint `gorm:"not null;index:idx_pending_user_priority" json:"user_id"`

	// Reference to the actual message
	MessageID uint    `gorm:"not null" json:"message_id"`
	Message   Message `gorm:"foreignKey:MessageID" json:"message"`

	// Delivery tracking
	Attempts    int        `gorm:"default:0" json:"attempts"`
	LastAttempt *time.Time `json:"last_attempt"`
	NextRetry   *time.Time `gorm:"index" json:"next_retry"` // For exponential backoff

	// Priority for message ordering (system messages can have higher priority)
	Priority int `gorm:"default:0;index:idx_pending_user_priority" json:"priority"`

	// Payload to send (cached JSON to avoid joins on delivery)
	Payload string `gorm:"type:text" json:"payload"`
}
