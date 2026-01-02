package models

import (
	"testing"
	"time"
)

func TestUserToResponse(t *testing.T) {
	now := time.Now()
	user := &User{
		ID:       1,
		Username: "john_doe",
		Email:    "john@example.com",
		FullName: "John Doe",
		Avatar:   "https://example.com/avatar.jpg",
		Role:     "user",
		IsOnline: true,
		LastSeen: &now,
	}

	response := user.ToResponse()

	if response.ID != user.ID {
		t.Errorf("ToResponse ID = %d, want %d", response.ID, user.ID)
	}
	if response.Username != user.Username {
		t.Errorf("ToResponse Username = %q, want %q", response.Username, user.Username)
	}
	if response.Email != user.Email {
		t.Errorf("ToResponse Email = %q, want %q", response.Email, user.Email)
	}
	if response.FullName != user.FullName {
		t.Errorf("ToResponse FullName = %q, want %q", response.FullName, user.FullName)
	}
	if response.Avatar != user.Avatar {
		t.Errorf("ToResponse Avatar = %q, want %q", response.Avatar, user.Avatar)
	}
	if response.Role != user.Role {
		t.Errorf("ToResponse Role = %q, want %q", response.Role, user.Role)
	}
	if response.IsOnline != user.IsOnline {
		t.Errorf("ToResponse IsOnline = %v, want %v", response.IsOnline, user.IsOnline)
	}
	if response.LastSeen == nil {
		t.Errorf("ToResponse LastSeen is nil")
	}
}

func TestMessageToResponse(t *testing.T) {
	createdAt := time.Now()
	senderID := uint(1)
	recipientID := uint(2)

	message := &Message{
		ID:          1,
		CreatedAt:   createdAt,
		ClientID:    "client-123",
		SenderID:    senderID,
		RecipientID: &recipientID,
		GroupID:     nil,
		Content:     "Hello, world!",
		MessageType: TextMessage,
		Status:      StatusSent,
		IsDelivered: true,
		IsRead:      false,
		Version:     1,
		Sender: User{
			ID:       senderID,
			Username: "john_doe",
			Email:    "john@example.com",
		},
	}

	response := message.ToResponse()

	if response.ID != message.ID {
		t.Errorf("ToResponse ID = %d, want %d", response.ID, message.ID)
	}
	if response.ClientID != message.ClientID {
		t.Errorf("ToResponse ClientID = %q, want %q", response.ClientID, message.ClientID)
	}
	if response.SenderID != message.SenderID {
		t.Errorf("ToResponse SenderID = %d, want %d", response.SenderID, message.SenderID)
	}
	if response.Content != message.Content {
		t.Errorf("ToResponse Content = %q, want %q", response.Content, message.Content)
	}
	if response.MessageType != message.MessageType {
		t.Errorf("ToResponse MessageType = %q, want %q", response.MessageType, message.MessageType)
	}
	if response.Status != message.Status {
		t.Errorf("ToResponse Status = %q, want %q", response.Status, message.Status)
	}
	if response.IsDelivered != message.IsDelivered {
		t.Errorf("ToResponse IsDelivered = %v, want %v", response.IsDelivered, message.IsDelivered)
	}
	if response.IsRead != message.IsRead {
		t.Errorf("ToResponse IsRead = %v, want %v", response.IsRead, message.IsRead)
	}
	if response.Version != message.Version {
		t.Errorf("ToResponse Version = %d, want %d", response.Version, message.Version)
	}
}

func TestMessageTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		msgType  MessageType
		expected string
	}{
		{"TextMessage", TextMessage, "text"},
		{"ImageMessage", ImageMessage, "image"},
		{"FileMessage", FileMessage, "file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.msgType) != tt.expected {
				t.Errorf("MessageType = %q, want %q", string(tt.msgType), tt.expected)
			}
		})
	}
}

func TestMessageStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		status   MessageStatus
		expected string
	}{
		{"StatusPending", StatusPending, "pending"},
		{"StatusSent", StatusSent, "sent"},
		{"StatusDelivered", StatusDelivered, "delivered"},
		{"StatusRead", StatusRead, "read"},
		{"StatusFailed", StatusFailed, "failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("MessageStatus = %q, want %q", string(tt.status), tt.expected)
			}
		})
	}
}

func TestUserResponseFields(t *testing.T) {
	response := UserResponse{
		ID:       1,
		Username: "john_doe",
		Email:    "john@example.com",
		FullName: "John Doe",
		Avatar:   "https://example.com/avatar.jpg",
		Role:     "user",
		IsOnline: true,
		LastSeen: nil,
	}

	if response.ID != 1 {
		t.Errorf("UserResponse ID = %d, want 1", response.ID)
	}
	if response.Username != "john_doe" {
		t.Errorf("UserResponse Username = %q, want john_doe", response.Username)
	}
	if response.IsOnline != true {
		t.Errorf("UserResponse IsOnline = %v, want true", response.IsOnline)
	}
	if response.LastSeen != nil {
		t.Errorf("UserResponse LastSeen = %v, want nil", response.LastSeen)
	}
}

func TestMessageResponseFields(t *testing.T) {
	createdAt := time.Now()
	response := MessageResponse{
		ID:          1,
		ClientID:    "client-123",
		SenderID:    1,
		Content:     "Test message",
		MessageType: TextMessage,
		Status:      StatusSent,
		IsDelivered: true,
		IsRead:      false,
		Version:     1,
		CreatedAt:   createdAt,
	}

	if response.ID != 1 {
		t.Errorf("MessageResponse ID = %d, want 1", response.ID)
	}
	if response.ClientID != "client-123" {
		t.Errorf("MessageResponse ClientID = %q, want client-123", response.ClientID)
	}
	if response.MessageType != TextMessage {
		t.Errorf("MessageResponse MessageType = %q, want text", response.MessageType)
	}
	if response.IsDelivered != true {
		t.Errorf("MessageResponse IsDelivered = %v, want true", response.IsDelivered)
	}
}
