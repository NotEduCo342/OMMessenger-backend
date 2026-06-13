package models

import (
	"time"
)

type DeletedConversation struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	UserID         uint      `gorm:"index:idx_user_convo,unique;not null" json:"user_id"`
	ConversationID string    `gorm:"type:varchar(50);index:idx_user_convo,unique;not null" json:"conversation_id"` // user_{id} or group_{id}
	ClearedAt      time.Time `gorm:"not null" json:"cleared_at"`
}
