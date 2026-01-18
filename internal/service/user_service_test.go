package service

import (
	"errors"
	"testing"

	"github.com/noteduco342/OMMessenger-backend/internal/models"
)

// MockUserRepository is a mock implementation of UserRepository for testing
type MockUserRepository struct {
	users  map[uint]*models.User
	nextID uint
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:  make(map[uint]*models.User),
		nextID: 1,
	}
}

func (m *MockUserRepository) Create(user *models.User) error {
	if user.ID == 0 {
		user.ID = m.nextID
		m.nextID++
	}
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) FindByEmail(email string) (*models.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, errors.New("record not found")
}

func (m *MockUserRepository) FindByUsername(username string) (*models.User, error) {
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, errors.New("record not found")
}

func (m *MockUserRepository) FindByID(id uint) (*models.User, error) {
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, errors.New("record not found")
}

func (m *MockUserRepository) Update(user *models.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) UpdateOnlineStatus(userID uint, isOnline bool) error {
	if user, ok := m.users[userID]; ok {
		user.IsOnline = isOnline
		return nil
	}
	return errors.New("record not found")
}

func (m *MockUserRepository) SearchUsers(query string, limit int) ([]models.User, error) {
	var results []models.User
	count := 0
	for _, user := range m.users {
		if count >= limit {
			break
		}
		results = append(results, *user)
		count++
	}
	return results, nil
}

// Tests for UserService

func TestIsUsernameAvailable(t *testing.T) {
	mockRepo := NewMockUserRepository()
	mockGroupRepo := NewMockGroupRepository()
	userService := NewUserService(mockRepo, mockGroupRepo)

	// Create a test user
	testUser := &models.User{
		Username: "existinguser",
		Email:    "test@example.com",
	}
	mockRepo.Create(testUser)

	tests := []struct {
		name      string
		username  string
		expected  bool
		shouldErr bool
	}{
		{"Available username", "newuser", true, false},
		{"Existing username", "existinguser", false, false},
		{"Empty username", "", false, true},
		{"Username with spaces", "  ", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := userService.IsUsernameAvailable(tt.username)
			if (err != nil) != tt.shouldErr {
				t.Errorf("IsUsernameAvailable(%q) error = %v, wantErr %v", tt.username, err, tt.shouldErr)
			}
			if result != tt.expected {
				t.Errorf("IsUsernameAvailable(%q) = %v, want %v", tt.username, result, tt.expected)
			}
		})
	}
}

func TestUpdateProfile(t *testing.T) {
	mockRepo := NewMockUserRepository()
	mockGroupRepo := NewMockGroupRepository()
	userService := NewUserService(mockRepo, mockGroupRepo)

	// Create a test user
	testUser := &models.User{
		ID:       1,
		Username: "john_doe",
		Email:    "john@example.com",
		FullName: "John Doe",
	}
	mockRepo.Create(testUser)

	tests := []struct {
		name      string
		userID    uint
		input     UpdateProfileInput
		expectErr bool
		checkFn   func(*models.User) bool
	}{
		{
			name:   "Update full name",
			userID: 1,
			input: UpdateProfileInput{
				FullName: "John Smith",
			},
			expectErr: false,
			checkFn: func(u *models.User) bool {
				return u.FullName == "John Smith"
			},
		},
		{
			name:   "Update username",
			userID: 1,
			input: UpdateProfileInput{
				Username: "john_smith",
			},
			expectErr: false,
			checkFn: func(u *models.User) bool {
				return u.Username == "john_smith"
			},
		},
		{
			name:   "Update to existing username",
			userID: 1,
			input: UpdateProfileInput{
				Username: "john_doe", // Same as current
			},
			expectErr: false,
			checkFn: func(u *models.User) bool {
				return u.Username == "john_doe"
			},
		},
		{
			name:      "User not found",
			userID:    999,
			input:     UpdateProfileInput{},
			expectErr: true,
			checkFn:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := userService.UpdateProfile(tt.userID, tt.input)
			if (err != nil) != tt.expectErr {
				t.Errorf("UpdateProfile error = %v, wantErr %v", err, tt.expectErr)
			}
			if !tt.expectErr && tt.checkFn != nil {
				if !tt.checkFn(result) {
					t.Errorf("UpdateProfile result does not match expected condition")
				}
			}
		})
	}
}

func TestGetUserByID(t *testing.T) {
	mockRepo := NewMockUserRepository()
	mockGroupRepo := NewMockGroupRepository()
	userService := NewUserService(mockRepo, mockGroupRepo)

	testUser := &models.User{
		ID:       1,
		Username: "john_doe",
		Email:    "john@example.com",
	}
	mockRepo.Create(testUser)

	tests := []struct {
		name      string
		userID    uint
		expectErr bool
	}{
		{"Existing user", 1, false},
		{"Non-existing user", 999, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := userService.GetUserByID(tt.userID)
			if (err != nil) != tt.expectErr {
				t.Errorf("GetUserByID error = %v, wantErr %v", err, tt.expectErr)
			}
			if !tt.expectErr && result.ID != tt.userID {
				t.Errorf("GetUserByID returned user with ID %d, want %d", result.ID, tt.userID)
			}
		})
	}
}

func TestGetUserByUsername(t *testing.T) {
	mockRepo := NewMockUserRepository()
	mockGroupRepo := NewMockGroupRepository()
	userService := NewUserService(mockRepo, mockGroupRepo)

	testUser := &models.User{
		ID:       1,
		Username: "john_doe",
		Email:    "john@example.com",
	}
	mockRepo.Create(testUser)

	tests := []struct {
		name      string
		username  string
		expectErr bool
	}{
		{"Existing username", "john_doe", false},
		{"Non-existing username", "nonexistent", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := userService.GetUserByUsername(tt.username)
			if (err != nil) != tt.expectErr {
				t.Errorf("GetUserByUsername error = %v, wantErr %v", err, tt.expectErr)
			}
			if !tt.expectErr && result.Username != tt.username {
				t.Errorf("GetUserByUsername returned user with username %q, want %q", result.Username, tt.username)
			}
		})
	}
}

func TestSearchUsers(t *testing.T) {
	mockRepo := NewMockUserRepository()
	mockGroupRepo := NewMockGroupRepository()
	userService := NewUserService(mockRepo, mockGroupRepo)

	// Create test users
	users := []*models.User{
		{ID: 1, Username: "john_doe", FullName: "John Doe"},
		{ID: 2, Username: "jane_smith", FullName: "Jane Smith"},
		{ID: 3, Username: "bob_jones", FullName: "Bob Jones"},
	}
	for _, u := range users {
		mockRepo.Create(u)
	}

	tests := []struct {
		name   string
		query  string
		limit  int
		minLen int
		maxLen int
	}{
		{"Search all", "", 10, 0, 10},
		{"Search with limit", "", 2, 0, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := userService.SearchUsers(tt.query, tt.limit)
			if err != nil {
				t.Errorf("SearchUsers error = %v", err)
			}
			if len(result) < tt.minLen || len(result) > tt.maxLen {
				t.Errorf("SearchUsers returned %d users, want between %d and %d", len(result), tt.minLen, tt.maxLen)
			}
		})
	}
}
