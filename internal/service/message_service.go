package service

import (
	"time"

	"github.com/noteduco342/OMMessenger-backend/internal/models"
	"github.com/noteduco342/OMMessenger-backend/internal/repository"
)

type MessageService struct {
	messageRepo repository.MessageRepositoryInterface
}

func NewMessageService(messageRepo repository.MessageRepositoryInterface) *MessageService {
	return &MessageService{messageRepo: messageRepo}
}

type SendMessageInput struct {
	RecipientID *uint              `json:"recipient_id"`
	GroupID     *uint              `json:"group_id"`
	Content     string             `json:"content"`
	MessageType models.MessageType `json:"message_type"`
}

func (s *MessageService) SendMessage(senderID uint, input SendMessageInput) (*models.Message, error) {
	message := &models.Message{
		SenderID:    senderID,
		RecipientID: input.RecipientID,
		GroupID:     input.GroupID,
		Content:     input.Content,
		MessageType: input.MessageType,
	}

	if message.MessageType == "" {
		message.MessageType = models.TextMessage
	}

	if err := s.messageRepo.Create(message); err != nil {
		return nil, err
	}

	// Load sender info
	return s.messageRepo.FindByID(message.ID)
}

func (s *MessageService) GetConversation(userID1, userID2 uint, limit int) ([]models.Message, error) {
	if limit == 0 {
		limit = 50
	}
	return s.messageRepo.FindConversation(userID1, userID2, limit)
}

// GetConversationCursor fetches messages with cursor-based pagination
func (s *MessageService) GetConversationCursor(userID1, userID2 uint, cursor uint, limit int) ([]models.Message, error) {
	if limit == 0 || limit > 100 {
		limit = 50
	}
	return s.messageRepo.FindConversationCursor(userID1, userID2, cursor, limit)
}

func (s *MessageService) MarkAsDelivered(messageID uint) error {
	return s.messageRepo.MarkAsDelivered(messageID)
}

func (s *MessageService) MarkAsRead(messageID uint) error {
	return s.messageRepo.MarkAsRead(messageID)
}

func (s *MessageService) GetByID(messageID uint) (*models.Message, error) {
	return s.messageRepo.FindByID(messageID)
}

func (s *MessageService) MarkConversationAsRead(userID uint, peerID uint) (int64, error) {
	return s.messageRepo.MarkConversationAsRead(userID, peerID)
}

// CreateWithClientID creates a message with client ID for deduplication
func (s *MessageService) CreateWithClientID(senderID uint, clientID string, recipientID *uint, groupID *uint, content string) (*models.Message, error) {
	message := &models.Message{
		ClientID:    clientID,
		SenderID:    senderID,
		RecipientID: recipientID,
		GroupID:     groupID,
		Content:     content,
		MessageType: models.TextMessage,
		Status:      models.StatusSent,
	}

	if err := s.messageRepo.Create(message); err != nil {
		return nil, err
	}

	return s.messageRepo.FindByID(message.ID)
}

// CreateWithClientIDAndType creates a message with client ID and message type for deduplication
func (s *MessageService) CreateWithClientIDAndType(senderID uint, clientID string, recipientID *uint, groupID *uint, content string, messageType models.MessageType) (*models.Message, error) {
	if messageType == "" {
		messageType = models.TextMessage
	}

	message := &models.Message{
		ClientID:    clientID,
		SenderID:    senderID,
		RecipientID: recipientID,
		GroupID:     groupID,
		Content:     content,
		MessageType: messageType,
		Status:      models.StatusSent,
	}

	if err := s.messageRepo.Create(message); err != nil {
		return nil, err
	}

	return s.messageRepo.FindByID(message.ID)
}

// GetByClientID finds a message by client ID and sender
func (s *MessageService) GetByClientID(clientID string, senderID uint) (*models.Message, error) {
	return s.messageRepo.FindByClientID(clientID, senderID)
}

// GetMessagesSince gets messages for a conversation since a specific message ID
func (s *MessageService) GetMessagesSince(requestingUserID uint, conversationID string, lastMessageID uint, limit int) ([]models.Message, error) {
	if limit == 0 || limit > 100 {
		limit = 100
	}
	return s.messageRepo.FindMessagesSince(requestingUserID, conversationID, lastMessageID, limit)
}

func (s *MessageService) GetGroupMessages(groupID uint, cursor uint, limit int) ([]models.Message, error) {
	if limit == 0 || limit > 100 {
		limit = 50
	}
	return s.messageRepo.FindGroupMessages(groupID, cursor, limit)
}

func (s *MessageService) ListGroupConversations(userID uint, cursorCreatedAt *time.Time, cursorMessageID uint, limit int) ([]repository.GroupConversationRow, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	return s.messageRepo.ListGroupConversations(userID, cursorCreatedAt, cursorMessageID, limit)
}

func (s *MessageService) ListConversationsUnified(userID uint, cursorCreatedAt *time.Time, cursorMessageID uint, limit int) ([]repository.ConversationUnifiedRow, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	return s.messageRepo.ListConversationsUnified(userID, cursorCreatedAt, cursorMessageID, limit)
}

func (s *MessageService) GetLatestGroupMessageID(groupID uint) (uint, error) {
	return s.messageRepo.GetLatestGroupMessageID(groupID)
}

func (s *MessageService) GetLatestDirectMessageID(userID1, userID2 uint) (uint, error) {
	return s.messageRepo.GetLatestDirectMessageID(userID1, userID2)
}

func (s *MessageService) IsMessageInGroup(messageID uint, groupID uint) (bool, error) {
	return s.messageRepo.IsMessageInGroup(messageID, groupID)
}

func (s *MessageService) ListDirectConversations(userID uint, cursorCreatedAt *time.Time, cursorMessageID uint, limit int) ([]repository.ConversationRow, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	return s.messageRepo.ListDirectConversations(userID, cursorCreatedAt, cursorMessageID, limit)
}
