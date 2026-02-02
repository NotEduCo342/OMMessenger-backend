package service

import (
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/noteduco342/OMMessenger-backend/internal/models"
	"github.com/noteduco342/OMMessenger-backend/internal/repository"
)

// MockMessageRepository is a mock implementation of MessageRepository for testing
type MockMessageRepository struct {
	messages map[uint]*models.Message
	nextID   uint
}

func NewMockMessageRepository() *MockMessageRepository {
	return &MockMessageRepository{
		messages: make(map[uint]*models.Message),
		nextID:   1,
	}
}

func (m *MockMessageRepository) Create(message *models.Message) error {
	if message.ID == 0 {
		message.ID = m.nextID
		m.nextID++
	}
	m.messages[message.ID] = message
	return nil
}

func (m *MockMessageRepository) FindByID(id uint) (*models.Message, error) {
	if msg, ok := m.messages[id]; ok {
		return msg, nil
	}
	return nil, errors.New("record not found")
}

func (m *MockMessageRepository) FindByClientID(clientID string, senderID uint) (*models.Message, error) {
	for _, msg := range m.messages {
		if msg.ClientID == clientID && msg.SenderID == senderID {
			return msg, nil
		}
	}
	return nil, errors.New("record not found")
}

func (m *MockMessageRepository) FindConversation(userID1, userID2 uint, limit int) ([]models.Message, error) {
	var result []models.Message
	count := 0
	for _, msg := range m.messages {
		if count >= limit {
			break
		}
		if (msg.SenderID == userID1 && msg.RecipientID != nil && *msg.RecipientID == userID2) ||
			(msg.SenderID == userID2 && msg.RecipientID != nil && *msg.RecipientID == userID1) {
			result = append(result, *msg)
			count++
		}
	}
	return result, nil
}

func (m *MockMessageRepository) FindConversationCursor(userID1, userID2 uint, cursor uint, limit int) ([]models.Message, error) {
	var result []models.Message
	count := 0
	for _, msg := range m.messages {
		if count >= limit {
			break
		}
		if msg.ID < cursor && ((msg.SenderID == userID1 && msg.RecipientID != nil && *msg.RecipientID == userID2) ||
			(msg.SenderID == userID2 && msg.RecipientID != nil && *msg.RecipientID == userID1)) {
			result = append(result, *msg)
			count++
		}
	}
	return result, nil
}

func (m *MockMessageRepository) FindGroupMessages(groupID uint, cursor uint, limit int) ([]models.Message, error) {
	var result []models.Message
	count := 0
	for _, msg := range m.messages {
		if count >= limit {
			break
		}
		if msg.GroupID == nil || *msg.GroupID != groupID {
			continue
		}
		if cursor > 0 && msg.ID >= cursor {
			continue
		}
		result = append(result, *msg)
		count++
	}
	return result, nil
}

func (m *MockMessageRepository) GetLatestDirectMessageID(userID1, userID2 uint) (uint, error) {
	var maxID uint
	for _, msg := range m.messages {
		if msg.GroupID != nil {
			continue
		}
		if (msg.SenderID == userID1 && msg.RecipientID != nil && *msg.RecipientID == userID2) ||
			(msg.SenderID == userID2 && msg.RecipientID != nil && *msg.RecipientID == userID1) {
			if msg.ID > maxID {
				maxID = msg.ID
			}
		}
	}
	return maxID, nil
}

func (m *MockMessageRepository) ListRecentPeers(userID uint, limit int) ([]repository.RecentPeerRow, error) {
	return []repository.RecentPeerRow{}, nil
}

func (m *MockMessageRepository) FindMessagesSince(requestingUserID uint, conversationID string, lastMessageID uint, limit int) ([]models.Message, error) {
	var result []models.Message
	count := 0
	for _, msg := range m.messages {
		if count >= limit {
			break
		}
		if msg.ID <= lastMessageID {
			continue
		}
		// Very small mock: support direct conversations of the form user_<id>
		if strings.HasPrefix(conversationID, "user_") {
			otherStr := strings.TrimPrefix(conversationID, "user_")
			otherID64, err := strconv.ParseUint(otherStr, 10, 32)
			if err != nil {
				return nil, err
			}
			otherID := uint(otherID64)
			if (msg.SenderID == requestingUserID && msg.RecipientID != nil && *msg.RecipientID == otherID) ||
				(msg.SenderID == otherID && msg.RecipientID != nil && *msg.RecipientID == requestingUserID) {
				result = append(result, *msg)
				count++
			}
		}
	}
	return result, nil
}

func (m *MockMessageRepository) ListDirectConversations(userID uint, cursorCreatedAt *time.Time, cursorMessageID uint, limit int) ([]repository.ConversationRow, error) {
	return []repository.ConversationRow{}, nil
}

func (m *MockMessageRepository) ListGroupConversations(userID uint, cursorCreatedAt *time.Time, cursorMessageID uint, limit int) ([]repository.GroupConversationRow, error) {
	return []repository.GroupConversationRow{}, nil
}

func (m *MockMessageRepository) ListConversationsUnified(userID uint, cursorCreatedAt *time.Time, cursorMessageID uint, limit int) ([]repository.ConversationUnifiedRow, error) {
	return []repository.ConversationUnifiedRow{}, nil
}

func (m *MockMessageRepository) GetLatestGroupMessageID(groupID uint) (uint, error) {
	var maxID uint
	for _, msg := range m.messages {
		if msg.GroupID != nil && *msg.GroupID == groupID {
			if msg.ID > maxID {
				maxID = msg.ID
			}
		}
	}
	return maxID, nil
}

func (m *MockMessageRepository) IsMessageInGroup(messageID uint, groupID uint) (bool, error) {
	if msg, ok := m.messages[messageID]; ok {
		return msg.GroupID != nil && *msg.GroupID == groupID, nil
	}
	return false, nil
}

func (m *MockMessageRepository) MarkAsDelivered(messageID uint) error {
	if msg, ok := m.messages[messageID]; ok {
		msg.IsDelivered = true
		msg.Status = models.StatusDelivered
		return nil
	}
	return errors.New("record not found")
}

func (m *MockMessageRepository) MarkAsRead(messageID uint) error {
	if msg, ok := m.messages[messageID]; ok {
		msg.IsRead = true
		msg.Status = models.StatusRead
		return nil
	}
	return errors.New("record not found")
}

func (m *MockMessageRepository) MarkConversationAsRead(userID uint, peerID uint) (int64, error) {
	// Minimal mock: mark messages from peer -> user as read.
	var cleared int64
	for _, msg := range m.messages {
		if msg.RecipientID != nil && *msg.RecipientID == userID && msg.SenderID == peerID && !msg.IsRead {
			msg.IsRead = true
			msg.Status = models.StatusRead
			cleared++
		}
	}
	return cleared, nil
}

// Tests for MessageService

func TestSendMessage(t *testing.T) {
	mockRepo := NewMockMessageRepository()
	messageService := NewMessageService(mockRepo)

	recipientID := uint(2)
	tests := []struct {
		name      string
		senderID  uint
		input     SendMessageInput
		shouldErr bool
		checkFn   func(*models.Message) bool
	}{
		{
			name:     "Send text message",
			senderID: 1,
			input: SendMessageInput{
				RecipientID: &recipientID,
				Content:     "Hello, world!",
				MessageType: models.TextMessage,
			},
			shouldErr: false,
			checkFn: func(m *models.Message) bool {
				return m.Content == "Hello, world!" && m.MessageType == models.TextMessage
			},
		},
		{
			name:     "Send message with default type",
			senderID: 1,
			input: SendMessageInput{
				RecipientID: &recipientID,
				Content:     "Default message",
			},
			shouldErr: false,
			checkFn: func(m *models.Message) bool {
				return m.MessageType == models.TextMessage
			},
		},
		{
			name:     "Send image message",
			senderID: 1,
			input: SendMessageInput{
				RecipientID: &recipientID,
				Content:     "image.jpg",
				MessageType: models.ImageMessage,
			},
			shouldErr: false,
			checkFn: func(m *models.Message) bool {
				return m.MessageType == models.ImageMessage
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := messageService.SendMessage(tt.senderID, tt.input)
			if (err != nil) != tt.shouldErr {
				t.Errorf("SendMessage error = %v, wantErr %v", err, tt.shouldErr)
			}
			if !tt.shouldErr && result == nil {
				t.Errorf("SendMessage returned nil message")
			}
			if !tt.shouldErr && tt.checkFn != nil && !tt.checkFn(result) {
				t.Errorf("SendMessage result does not match expected condition")
			}
		})
	}
}

func TestGetConversation(t *testing.T) {
	mockRepo := NewMockMessageRepository()
	messageService := NewMessageService(mockRepo)

	recipientID := uint(2)
	// Create test messages
	mockRepo.Create(&models.Message{
		SenderID:    1,
		RecipientID: &recipientID,
		Content:     "Message 1",
	})
	recipientID2 := uint(1)
	mockRepo.Create(&models.Message{
		SenderID:    2,
		RecipientID: &recipientID2,
		Content:     "Message 2",
	})

	tests := []struct {
		name      string
		userID1   uint
		userID2   uint
		limit     int
		shouldErr bool
		minCount  int
	}{
		{"Get conversation", 1, 2, 50, false, 2},
		{"Get conversation with limit", 1, 2, 1, false, 0},
		{"Get conversation default limit", 1, 2, 0, false, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := messageService.GetConversation(tt.userID1, tt.userID2, tt.limit)
			if (err != nil) != tt.shouldErr {
				t.Errorf("GetConversation error = %v, wantErr %v", err, tt.shouldErr)
			}
			if !tt.shouldErr && len(result) < tt.minCount {
				t.Errorf("GetConversation returned %d messages, want at least %d", len(result), tt.minCount)
			}
		})
	}
}

func TestMarkAsDelivered(t *testing.T) {
	mockRepo := NewMockMessageRepository()
	messageService := NewMessageService(mockRepo)

	recipientID := uint(2)
	mockRepo.Create(&models.Message{
		ID:          1,
		SenderID:    1,
		RecipientID: &recipientID,
		Content:     "Test message",
		IsDelivered: false,
	})

	tests := []struct {
		name      string
		messageID uint
		shouldErr bool
	}{
		{"Mark existing message as delivered", 1, false},
		{"Mark non-existing message", 999, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := messageService.MarkAsDelivered(tt.messageID)
			if (err != nil) != tt.shouldErr {
				t.Errorf("MarkAsDelivered error = %v, wantErr %v", err, tt.shouldErr)
			}
			if !tt.shouldErr {
				msg, _ := mockRepo.FindByID(tt.messageID)
				if !msg.IsDelivered {
					t.Errorf("Message not marked as delivered")
				}
			}
		})
	}
}

func TestMarkAsRead(t *testing.T) {
	mockRepo := NewMockMessageRepository()
	messageService := NewMessageService(mockRepo)

	recipientID := uint(2)
	mockRepo.Create(&models.Message{
		ID:          1,
		SenderID:    1,
		RecipientID: &recipientID,
		Content:     "Test message",
		IsRead:      false,
	})

	tests := []struct {
		name      string
		messageID uint
		shouldErr bool
	}{
		{"Mark existing message as read", 1, false},
		{"Mark non-existing message", 999, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := messageService.MarkAsRead(tt.messageID)
			if (err != nil) != tt.shouldErr {
				t.Errorf("MarkAsRead error = %v, wantErr %v", err, tt.shouldErr)
			}
			if !tt.shouldErr {
				msg, _ := mockRepo.FindByID(tt.messageID)
				if !msg.IsRead {
					t.Errorf("Message not marked as read")
				}
			}
		})
	}
}

func TestCreateWithClientID(t *testing.T) {
	mockRepo := NewMockMessageRepository()
	messageService := NewMessageService(mockRepo)

	tests := []struct {
		name        string
		senderID    uint
		clientID    string
		recipientID *uint
		groupID     *uint
		content     string
		shouldErr   bool
	}{
		{
			name:        "Create message with client ID",
			senderID:    1,
			clientID:    "client-123",
			recipientID: &[]uint{2}[0],
			groupID:     nil,
			content:     "Test message",
			shouldErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := messageService.CreateWithClientID(tt.senderID, tt.clientID, tt.recipientID, tt.groupID, tt.content)
			if (err != nil) != tt.shouldErr {
				t.Errorf("CreateWithClientID error = %v, wantErr %v", err, tt.shouldErr)
			}
			if !tt.shouldErr && result.ClientID != tt.clientID {
				t.Errorf("CreateWithClientID returned message with clientID %q, want %q", result.ClientID, tt.clientID)
			}
		})
	}
}

func TestGetByClientID(t *testing.T) {
	mockRepo := NewMockMessageRepository()
	messageService := NewMessageService(mockRepo)

	recipientID := uint(2)
	mockRepo.Create(&models.Message{
		ClientID:    "client-123",
		SenderID:    1,
		RecipientID: &recipientID,
		Content:     "Test message",
	})

	tests := []struct {
		name      string
		clientID  string
		senderID  uint
		shouldErr bool
	}{
		{"Get existing message", "client-123", 1, false},
		{"Get non-existing message", "non-existent", 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := messageService.GetByClientID(tt.clientID, tt.senderID)
			if (err != nil) != tt.shouldErr {
				t.Errorf("GetByClientID error = %v, wantErr %v", err, tt.shouldErr)
			}
			if !tt.shouldErr && result.ClientID != tt.clientID {
				t.Errorf("GetByClientID returned message with clientID %q, want %q", result.ClientID, tt.clientID)
			}
		})
	}
}

func TestGetMessagesSince(t *testing.T) {
	mockRepo := NewMockMessageRepository()
	messageService := NewMessageService(mockRepo)

	requestingUserID := uint(1)
	otherUserID := uint(2)
	conversationID := "user_2"

	// Create test messages between user 1 and user 2
	mockRepo.Create(&models.Message{ID: 1, SenderID: requestingUserID, RecipientID: &otherUserID, Content: "Message 1"})
	mockRepo.Create(&models.Message{ID: 2, SenderID: otherUserID, RecipientID: &requestingUserID, Content: "Message 2"})
	mockRepo.Create(&models.Message{ID: 3, SenderID: requestingUserID, RecipientID: &otherUserID, Content: "Message 3"})

	tests := []struct {
		name             string
		conversationID   string
		lastMessageID    uint
		limit            int
		shouldErr        bool
		expectedMinCount int
	}{
		{"Get messages since ID 1", conversationID, 1, 50, false, 2},
		{"Get with limit", conversationID, 0, 1, false, 1},
		{"Get with excessive limit", conversationID, 0, 150, false, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := messageService.GetMessagesSince(requestingUserID, tt.conversationID, tt.lastMessageID, tt.limit)
			if (err != nil) != tt.shouldErr {
				t.Errorf("GetMessagesSince error = %v, wantErr %v", err, tt.shouldErr)
			}
			if !tt.shouldErr && len(result) < tt.expectedMinCount {
				t.Errorf("GetMessagesSince returned %d messages, want at least %d", len(result), tt.expectedMinCount)
			}
		})
	}
}
