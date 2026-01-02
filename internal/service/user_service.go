package service

import (
	"errors"
	"strings"

	"github.com/noteduco342/OMMessenger-backend/internal/models"
	"github.com/noteduco342/OMMessenger-backend/internal/repository"
)

type UserService struct {
	userRepo repository.UserRepositoryInterface
}

func NewUserService(userRepo repository.UserRepositoryInterface) *UserService {
	return &UserService{userRepo: userRepo}
}

type UpdateProfileInput struct {
	Username string `json:"username"`
	FullName string `json:"full_name"`
}

func (s *UserService) IsUsernameAvailable(username string) (bool, error) {
	// Normalize username
	username = strings.TrimSpace(username)
	if username == "" {
		return false, errors.New("username cannot be empty")
	}

	// Check if username exists
	_, err := s.userRepo.FindByUsername(username)
	if err != nil {
		// Username not found = available
		return true, nil
	}

	// Username found = not available
	return false, nil
}

func (s *UserService) UpdateProfile(userID uint, input UpdateProfileInput) (*models.User, error) {
	// Get current user
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Validate and update username if provided
	if input.Username != "" {
		username := strings.TrimSpace(input.Username)

		// Only check availability if username is different
		if username != user.Username {
			// Check if new username is available
			available, err := s.IsUsernameAvailable(username)
			if err != nil {
				return nil, err
			}
			if !available {
				return nil, errors.New("username already taken")
			}
			user.Username = username
		}
	}

	// Update full name if provided
	if input.FullName != "" {
		user.FullName = strings.TrimSpace(input.FullName)
	}

	// Save changes
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUserByID(userID uint) (*models.User, error) {
	return s.userRepo.FindByID(userID)
}

func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	username = strings.TrimSpace(strings.ToLower(username))
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}
	return s.userRepo.FindByUsername(username)
}

func (s *UserService) SearchUsers(query string, limit int) ([]models.User, error) {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return []models.User{}, nil
	}
	if limit == 0 || limit > 50 {
		limit = 20
	}
	return s.userRepo.SearchUsers(query, limit)
}
