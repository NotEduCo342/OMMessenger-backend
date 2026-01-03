package models

import (
	"time"

	"gorm.io/gorm"
)

type GroupRole string

const (
	RoleAdmin  GroupRole = "admin"
	RoleMember GroupRole = "member"
)

type Group struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name        string `gorm:"size:100;not null" json:"name"`
	Description string `gorm:"size:255" json:"description"`
	Icon        string `json:"icon"`
	CreatorID   uint   `gorm:"not null" json:"creator_id"`

	// Associations
	Creator User          `gorm:"foreignKey:CreatorID" json:"creator"`
	Members []GroupMember `gorm:"foreignKey:GroupID" json:"members"`
}

type GroupMember struct {
	GroupID  uint      `gorm:"primaryKey" json:"group_id"`
	UserID   uint      `gorm:"primaryKey" json:"user_id"`
	Role     GroupRole `gorm:"type:varchar(20);default:'member'" json:"role"`
	JoinedAt time.Time `gorm:"autoCreateTime" json:"joined_at"`

	User  User  `gorm:"foreignKey:UserID" json:"user"`
	Group Group `gorm:"foreignKey:GroupID" json:"-"`
}
