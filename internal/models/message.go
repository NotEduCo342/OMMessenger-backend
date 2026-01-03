package models

import (
	"time"

	"gorm.io/gorm"
)

type MessageType string

const (
	TextMessage  MessageType = "text"
	ImageMessage MessageType = "image"
	FileMessage  MessageType = "file"
)

type MessageStatus string

const (
	StatusPending   MessageStatus = "pending"
	StatusSent      MessageStatus = "sent"
	StatusDelivered MessageStatus = "delivered"
	StatusRead      MessageStatus = "read"
	StatusFailed    MessageStatus = "failed"
)

type Message struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Client-side tracking
	ClientID string `gorm:"type:varchar(36);uniqueIndex:idx_client_sender;not null" json:"client_id"` // UUID for deduplication

	SenderID    uint   `gorm:"not null;uniqueIndex:idx_client_sender;index" json:"sender_id"`
	Sender      User   `gorm:"foreignKey:SenderID" json:"sender"`
	RecipientID *uint  `gorm:"index" json:"recipient_id"` // null for group messages
	GroupID     *uint  `gorm:"index" json:"group_id"`     // null for direct messages
	Group       *Group `gorm:"foreignKey:GroupID" json:"group,omitempty"`

	Content     string      `gorm:"type:text;not null" json:"content"`
	MessageType MessageType `gorm:"type:varchar(20);default:'text'" json:"message_type"`

	// Status tracking
	Status      MessageStatus `gorm:"type:varchar(20);default:'pending';index" json:"status"`
	IsDelivered bool          `gorm:"default:false" json:"is_delivered"`
	IsRead      bool          `gorm:"default:false" json:"is_read"`
	DeliveredAt *time.Time    `json:"delivered_at"`
	ReadAt      *time.Time    `json:"read_at"`

	// Version for edit tracking
	Version int `gorm:"default:1" json:"version"`

	// For encryption (optional)
	IsEncrypted bool `gorm:"default:false" json:"is_encrypted"`
}

type MessageResponse struct {
	ID          uint          `json:"id"`
	ClientID    string        `json:"client_id"`
	SenderID    uint          `json:"sender_id"`
	Sender      UserResponse  `json:"sender"`
	RecipientID *uint         `json:"recipient_id"`
	GroupID     *uint         `json:"group_id"`
	Content     string        `json:"content"`
	MessageType MessageType   `json:"message_type"`
	Status      MessageStatus `json:"status"`
	IsDelivered bool          `json:"is_delivered"`
	IsRead      bool          `json:"is_read"`
	Version     int           `json:"version"`
	CreatedAt   time.Time     `json:"created_at"`
}

func (m *Message) ToResponse() MessageResponse {
	return MessageResponse{
		ID:          m.ID,
		ClientID:    m.ClientID,
		SenderID:    m.SenderID,
		Sender:      m.Sender.ToResponse(),
		RecipientID: m.RecipientID,
		GroupID:     m.GroupID,
		Content:     m.Content,
		MessageType: m.MessageType,
		Status:      m.Status,
		IsDelivered: m.IsDelivered,
		IsRead:      m.IsRead,
		Version:     m.Version,
		CreatedAt:   m.CreatedAt,
	}
}
