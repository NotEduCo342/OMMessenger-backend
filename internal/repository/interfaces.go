package repository

import (
	"time"

	"github.com/noteduco342/OMMessenger-backend/internal/models"
)

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
	FindConversationCursor(userID1, userID2 uint, cursor uint, limit int) ([]models.Message, error)
	FindMessagesSince(requestingUserID uint, conversationID string, lastMessageID uint, limit int) ([]models.Message, error)
	ListDirectConversations(userID uint, cursorCreatedAt *time.Time, cursorMessageID uint, limit int) ([]ConversationRow, error)
	MarkAsDelivered(messageID uint) error
	MarkAsRead(messageID uint) error
	MarkConversationAsRead(userID uint, peerID uint) (int64, error)
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

// PendingMessageRepositoryInterface defines the contract for pending message queue operations
type PendingMessageRepositoryInterface interface {
	Enqueue(userID, messageID uint, payload string, priority int) error
	GetPendingForUser(userID uint, limit int) ([]models.PendingMessage, error)
	GetRetryable(limit int) ([]models.PendingMessage, error)
	MarkAttempted(id uint, attempts int, nextRetry *time.Time) error
	Delete(id uint) error
	DeleteBatch(ids []uint) error
	CountPendingForUser(userID uint) (int64, error)
	CleanupOld(olderThan time.Duration) error
}
