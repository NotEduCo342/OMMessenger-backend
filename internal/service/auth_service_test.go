package service

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/noteduco342/OMMessenger-backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// MockRefreshTokenRepository is a mock implementation for testing
type MockRefreshTokenRepository struct {
	tokens map[string]*models.RefreshToken
}

func NewMockRefreshTokenRepository() *MockRefreshTokenRepository {
	return &MockRefreshTokenRepository{
		tokens: make(map[string]*models.RefreshToken),
	}
}

func (m *MockRefreshTokenRepository) Create(token *models.RefreshToken) error {
	m.tokens[token.TokenHash] = token
	return nil
}

func (m *MockRefreshTokenRepository) FindValidByHash(hash string) (*models.RefreshToken, error) {
	token, ok := m.tokens[hash]
	if !ok {
		return nil, errors.New("record not found")
	}
	if time.Now().After(token.ExpiresAt) {
		return nil, errors.New("token expired")
	}
	return token, nil
}

func (m *MockRefreshTokenRepository) RevokeByHash(hash string) error {
	delete(m.tokens, hash)
	return nil
}

// Tests for AuthService

func TestRegister(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key-12345")

	mockUserRepo := NewMockUserRepository()
	mockRefreshTokenRepo := NewMockRefreshTokenRepository()
	mockGroupRepo := NewMockGroupRepository()
	authService := NewAuthService(mockUserRepo, mockRefreshTokenRepo, mockGroupRepo)

	tests := []struct {
		name      string
		input     RegisterInput
		shouldErr bool
	}{
		{
			name: "Valid registration",
			input: RegisterInput{
				Username: "john_doe",
				Email:    "john@example.com",
				Password: "securepassword123",
				FullName: "John Doe",
			},
			shouldErr: false,
		},
		{
			name: "Duplicate email",
			input: RegisterInput{
				Username: "jane_doe",
				Email:    "duplicate@example.com",
				Password: "securepassword123",
				FullName: "Jane Doe",
			},
			shouldErr: true,
		},
		{
			name: "Duplicate username",
			input: RegisterInput{
				Username: "duplicate_user",
				Email:    "another@example.com",
				Password: "securepassword123",
				FullName: "Another User",
			},
			shouldErr: true,
		},
	}

	// Pre-populate duplicate data
	mockUserRepo.Create(&models.User{
		Username: "duplicate_user",
		Email:    "duplicate@example.com",
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := authService.Register(tt.input)
			if (err != nil) != tt.shouldErr {
				t.Errorf("Register error = %v, wantErr %v", err, tt.shouldErr)
			}
			if !tt.shouldErr && result == nil {
				t.Errorf("Register returned nil AuthSession")
			}
			if !tt.shouldErr && result.AccessToken == "" {
				t.Errorf("Register returned empty access token")
			}
			if !tt.shouldErr && result.RefreshToken == "" {
				t.Errorf("Register returned empty refresh token")
			}
		})
	}
}

func TestLogin(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key-12345")

	mockUserRepo := NewMockUserRepository()
	mockRefreshTokenRepo := NewMockRefreshTokenRepository()
	mockGroupRepo := NewMockGroupRepository()
	authService := NewAuthService(mockUserRepo, mockRefreshTokenRepo, mockGroupRepo)

	// Create a test user with hashed password
	testPassword := "securepassword123"
	hashedPassword, _ := hashPassword(testPassword)
	testUser := &models.User{
		ID:           1,
		Username:     "john_doe",
		Email:        "john@example.com",
		PasswordHash: hashedPassword,
	}
	mockUserRepo.Create(testUser)

	tests := []struct {
		name      string
		input     LoginInput
		shouldErr bool
	}{
		{
			name: "Valid login",
			input: LoginInput{
				Email:    "john@example.com",
				Password: "securepassword123",
			},
			shouldErr: false,
		},
		{
			name: "Invalid email",
			input: LoginInput{
				Email:    "nonexistent@example.com",
				Password: "securepassword123",
			},
			shouldErr: true,
		},
		{
			name: "Wrong password",
			input: LoginInput{
				Email:    "john@example.com",
				Password: "wrongpassword",
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := authService.Login(tt.input)
			if (err != nil) != tt.shouldErr {
				t.Errorf("Login error = %v, wantErr %v", err, tt.shouldErr)
			}
			if !tt.shouldErr && result == nil {
				t.Errorf("Login returned nil AuthSession")
			}
		})
	}
}

func TestRefreshSession(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key-12345")

	mockUserRepo := NewMockUserRepository()
	mockRefreshTokenRepo := NewMockRefreshTokenRepository()
	mockGroupRepo := NewMockGroupRepository()
	authService := NewAuthService(mockUserRepo, mockRefreshTokenRepo, mockGroupRepo)

	testUser := &models.User{
		ID:       1,
		Email:    "john@example.com",
		Username: "john_doe",
		Role:     "user",
	}
	mockUserRepo.Create(testUser)

	// Generate a valid refresh token
	rawToken, tokenHash, _ := generateRefreshToken()
	refreshTokenModel := &models.RefreshToken{
		UserID:    1,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	mockRefreshTokenRepo.Create(refreshTokenModel)

	tests := []struct {
		name      string
		token     string
		shouldErr bool
	}{
		{"Valid refresh token", rawToken, false},
		{"Invalid refresh token", "invalid-token", true},
		{"Empty refresh token", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := authService.RefreshSession(tt.token)
			if (err != nil) != tt.shouldErr {
				t.Errorf("RefreshSession error = %v, wantErr %v", err, tt.shouldErr)
			}
			if !tt.shouldErr && result == nil {
				t.Errorf("RefreshSession returned nil AuthSession")
			}
		})
	}
}

func TestLogout(t *testing.T) {
	mockUserRepo := NewMockUserRepository()
	mockRefreshTokenRepo := NewMockRefreshTokenRepository()
	mockGroupRepo := NewMockGroupRepository()
	authService := NewAuthService(mockUserRepo, mockRefreshTokenRepo, mockGroupRepo)

	// Create a refresh token
	rawToken, tokenHash, _ := generateRefreshToken()
	refreshTokenModel := &models.RefreshToken{
		UserID:    1,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	mockRefreshTokenRepo.Create(refreshTokenModel)

	tests := []struct {
		name      string
		token     string
		shouldErr bool
	}{
		{"Valid logout", rawToken, false},
		{"Empty token", "", false},
		{"Non-existent token", "non-existent-token", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := authService.Logout(tt.token)
			if (err != nil) != tt.shouldErr {
				t.Errorf("Logout error = %v, wantErr %v", err, tt.shouldErr)
			}
		})
	}
}

// Helper function to hash password for testing
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
