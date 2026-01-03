package repository

import "github.com/noteduco342/OMMessenger-backend/internal/models"

// UserRepositoryInterface defines the contract for user repository operations
type UserRepositoryInterface interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByUsername(username string) (*models.User, error)
	FindByID(id uint) (*models.User, error)
	Update(user *models.User) error
	UpdateOnlineStatus(userID uint, isOnline bool) error
	SearchUsers(query string, limit int) ([]models.User, error)
}

// MessageRepositoryInterface defines the contract for message repository operations
type MessageRepositoryInterface interface {
	Create(message *models.Message) error
	FindByID(id uint) (*models.Message, error)
	FindByClientID(clientID string, senderID uint) (*models.Message, error)
	FindConversation(userID1, userID2 uint, limit int) ([]models.Message, error)
	FindMessagesSince(conversationID string, lastMessageID uint, limit int) ([]models.Message, error)
	MarkAsDelivered(messageID uint) error
	MarkAsRead(messageID uint) error
}

// RefreshTokenRepositoryInterface defines the contract for refresh token repository operations
type RefreshTokenRepositoryInterface interface {
	Create(token *models.RefreshToken) error
	FindValidByHash(tokenHash string) (*models.RefreshToken, error)
	RevokeByHash(tokenHash string) error
}

// GroupRepositoryInterface defines the contract for group repository operations
type GroupRepositoryInterface interface {
	Create(group *models.Group) error
	FindByID(id uint) (*models.Group, error)
	AddMember(groupID, userID uint, role models.GroupRole) error
	RemoveMember(groupID, userID uint) error
	GetMembers(groupID uint) ([]models.User, error)
	IsMember(groupID, userID uint) (bool, error)
	GetUserGroups(userID uint) ([]models.Group, error)
}
