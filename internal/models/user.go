package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Username     string     `gorm:"uniqueIndex;not null" json:"username"`
	Email        string     `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string     `gorm:"not null" json:"-"`
	FullName     string     `json:"full_name"`
	Avatar       string     `json:"avatar"`
	Role         string     `gorm:"not null;default:user" json:"role"`
	IsOnline     bool       `gorm:"default:false" json:"is_online"`
	LastSeen     *time.Time `json:"last_seen"`

	Messages []Message `gorm:"foreignKey:SenderID" json:"-"`
}

type UserResponse struct {
	ID       uint       `json:"id"`
	Username string     `json:"username"`
	Email    string     `json:"email"`
	FullName string     `json:"full_name"`
	Avatar   string     `json:"avatar"`
	Role     string     `json:"role"`
	IsOnline bool       `json:"is_online"`
	LastSeen *time.Time `json:"last_seen"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		FullName: u.FullName,
		Avatar:   u.Avatar,
		Role:     u.Role,
		IsOnline: u.IsOnline,
		LastSeen: u.LastSeen,
	}
}
