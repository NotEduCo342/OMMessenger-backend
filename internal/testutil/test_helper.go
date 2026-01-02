package testutil

import (
	"os"
	"testing"
	"time"

	"github.com/noteduco342/OMMessenger-backend/internal/models"
	"gorm.io/gorm"
)

// TestHelper provides utility functions for tests
type TestHelper struct {
	t *testing.T
}

func NewTestHelper(t *testing.T) *TestHelper {
	return &TestHelper{t: t}
}

// CreateTestUser creates a test user with default values
func (h *TestHelper) CreateTestUser(id uint, username, email string) *models.User {
	if id == 0 {
		id = 1
	}
	if username == "" {
		username = "testuser"
	}
	if email == "" {
		email = "test@example.com"
	}

	return &models.User{
		ID:           id,
		Username:     username,
		Email:        email,
		PasswordHash: "hashed_password_123",
		FullName:     "Test User",
		Avatar:       "https://example.com/avatar.jpg",
		Role:         "user",
		IsOnline:     false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// CreateTestMessage creates a test message with default values
func (h *TestHelper) CreateTestMessage(id uint, senderID uint, content string) *models.Message {
	if id == 0 {
		id = 1
	}
	if senderID == 0 {
		senderID = 1
	}
	if content == "" {
		content = "Test message"
	}

	recipientID := uint(2)
	return &models.Message{
		ID:          id,
		ClientID:    "client-" + string(rune(id)),
		SenderID:    senderID,
		RecipientID: &recipientID,
		Content:     content,
		MessageType: models.TextMessage,
		Status:      models.StatusSent,
		IsDelivered: false,
		IsRead:      false,
		Version:     1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Sender: models.User{
			ID:       senderID,
			Username: "sender",
			Email:    "sender@example.com",
		},
	}
}

// SetupTestEnv sets up required environment variables for testing
func (h *TestHelper) SetupTestEnv() {
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing-only")
	os.Setenv("DATABASE_URL", "")
	os.Setenv("PASSWORD_MIN_LENGTH", "10")
}

// TeardownTestEnv cleans up environment variables after testing
func (h *TestHelper) TeardownTestEnv() {
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("PASSWORD_MIN_LENGTH")
}

// AssertError checks if an error occurred when it should (or shouldn't)
func (h *TestHelper) AssertError(err error, shouldErr bool, testName string) {
	if (err != nil) != shouldErr {
		if shouldErr {
			h.t.Errorf("%s: expected error but got nil", testName)
		} else {
			h.t.Errorf("%s: unexpected error: %v", testName, err)
		}
	}
}

// AssertEqual checks if two values are equal
func (h *TestHelper) AssertEqual(got, want interface{}, testName string) {
	if got != want {
		h.t.Errorf("%s: got %v, want %v", testName, got, want)
	}
}

// AssertNotNil checks if a value is not nil
func (h *TestHelper) AssertNotNil(value interface{}, testName string) {
	if value == nil {
		h.t.Errorf("%s: expected non-nil value", testName)
	}
}

// AssertNil checks if a value is nil
func (h *TestHelper) AssertNil(value interface{}, testName string) {
	if value != nil {
		h.t.Errorf("%s: expected nil value but got %v", testName, value)
	}
}

// GetErrorWithMessage returns an error that mimics gorm.ErrRecordNotFound
func GetRecordNotFoundError() error {
	return gorm.ErrRecordNotFound
}
