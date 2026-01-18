package models

import (
	"time"
)

// GroupReadState tracks per-user read progress in a group.
// last_read_message_id is monotonic and represents the highest message ID the user has read.
type GroupReadState struct {
	GroupID           uint      `gorm:"primaryKey" json:"group_id"`
	UserID            uint      `gorm:"primaryKey" json:"user_id"`
	LastReadMessageID uint      `gorm:"not null;default:0" json:"last_read_message_id"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
