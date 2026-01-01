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

type Message struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	
	SenderID     uint        `gorm:"not null;index" json:"sender_id"`
	Sender       User        `gorm:"foreignKey:SenderID" json:"sender"`
	RecipientID  *uint       `gorm:"index" json:"recipient_id"` // null for group messages
	GroupID      *uint       `gorm:"index" json:"group_id"`     // null for direct messages
	
	Content      string      `gorm:"type:text;not null" json:"content"`
	MessageType  MessageType `gorm:"type:varchar(20);default:'text'" json:"message_type"`
	
	IsDelivered  bool       `gorm:"default:false" json:"is_delivered"`
	IsRead       bool       `gorm:"default:false" json:"is_read"`
	DeliveredAt  *time.Time `json:"delivered_at"`
	ReadAt       *time.Time `json:"read_at"`
	
	// For encryption (optional)
	IsEncrypted  bool   `gorm:"default:false" json:"is_encrypted"`
}

type MessageResponse struct {
	ID          uint         `json:"id"`
	SenderID    uint         `json:"sender_id"`
	Sender      UserResponse `json:"sender"`
	RecipientID *uint        `json:"recipient_id"`
	GroupID     *uint        `json:"group_id"`
	Content     string       `json:"content"`
	MessageType MessageType  `json:"message_type"`
	IsDelivered bool         `json:"is_delivered"`
	IsRead      bool         `json:"is_read"`
	CreatedAt   time.Time    `json:"created_at"`
}

func (m *Message) ToResponse() MessageResponse {
	return MessageResponse{
		ID:          m.ID,
		SenderID:    m.SenderID,
		Sender:      m.Sender.ToResponse(),
		RecipientID: m.RecipientID,
		GroupID:     m.GroupID,
		Content:     m.Content,
		MessageType: m.MessageType,
		IsDelivered: m.IsDelivered,
		IsRead:      m.IsRead,
		CreatedAt:   m.CreatedAt,
	}
}
