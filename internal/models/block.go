package models

import (
	"time"
)

type Block struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	BlockerID uint      `gorm:"uniqueIndex:idx_blocker_blocked;not null" json:"blocker_id"`
	BlockedID uint      `gorm:"uniqueIndex:idx_blocker_blocked;not null" json:"blocked_id"`

	Blocker User `gorm:"foreignKey:BlockerID" json:"-"`
	Blocked User `gorm:"foreignKey:BlockedID" json:"-"`
}
