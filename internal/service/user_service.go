package service

import (
	"errors"
	"strings"

	"github.com/noteduco342/OMMessenger-backend/internal/models"
	"github.com/noteduco342/OMMessenger-backend/internal/repository"
)

type UserService struct {
	userRepo  repository.UserRepositoryInterface
	groupRepo repository.GroupRepositoryInterface
}

func NewUserService(userRepo repository.UserRepositoryInterface, groupRepo repository.GroupRepositoryInterface) *UserService {
	return &UserService{userRepo: userRepo, groupRepo: groupRepo}
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
		// Username not found in users; check group handles
		if s.groupRepo != nil {
			if _, gErr := s.groupRepo.FindByHandle(username); gErr == nil {
				return false, nil
			}
		}
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

		// Only check availability if username is different (case-insensitive)
		if !strings.EqualFold(username, user.Username) {
			// Check if new username is available
			available, err := s.IsUsernameAvailable(username)
			if err != nil {
				return nil, err
			}
			if !available {
				return nil, errors.New("username already taken")
			}
			user.Username = username
		} else if username != user.Username {
			// Allow casing-only changes without availability check.
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

func (s *UserService) SetUserOnline(userID uint) error {
	return s.userRepo.UpdateOnlineStatus(userID, true)
}

func (s *UserService) SetUserOffline(userID uint) error {
	return s.userRepo.UpdateOnlineStatus(userID, false)
}

func (s *UserService) BlockUser(blockerID, blockedID uint) error {
	isBlocked, err := s.userRepo.IsBlocker(blockerID, blockedID)
	if err != nil {
		return err
	}
	if isBlocked {
		return errors.New("user is already blocked")
	}

	_, err = s.userRepo.FindByID(blockedID)
	if err != nil {
		return errors.New("user to block not found")
	}

	return s.userRepo.BlockUser(blockerID, blockedID)
}

func (s *UserService) UnblockUser(blockerID, blockedID uint) error {
	isBlocked, err := s.userRepo.IsBlocker(blockerID, blockedID)
	if err != nil {
		return err
	}
	if !isBlocked {
		return errors.New("user is not blocked")
	}

	return s.userRepo.UnblockUser(blockerID, blockedID)
}

func (s *UserService) GetBlockedUsers(userID uint) ([]models.User, error) {
	return s.userRepo.GetBlockedUsers(userID)
}

func (s *UserService) IsBlocked(userID1, userID2 uint) (bool, error) {
	return s.userRepo.IsBlocked(userID1, userID2)
}

func (s *UserService) IsBlocker(blockerID, blockedID uint) (bool, error) {
	return s.userRepo.IsBlocker(blockerID, blockedID)
}

func (s *UserService) GetBlockedUserIDs(userID uint) ([]uint, error) {
	return s.userRepo.GetBlockerIDs(userID)
}

func (s *UserService) GetBlockRelationshipsForUser(userID uint) ([]uint, error) {
	return s.userRepo.GetBlockRelationshipsForUser(userID)
}

func (s *UserService) GetBlockMaps(userID uint) (blockedByMe map[uint]bool, blockedByPeer map[uint]bool, err error) {
	blockers, err := s.userRepo.GetBlockerIDs(userID)
	if err != nil {
		return nil, nil, err
	}

	blockedList, err := s.userRepo.GetBlockedUsers(userID)
	if err != nil {
		return nil, nil, err
	}

	blockedByMe = make(map[uint]bool)
	for _, u := range blockedList {
		blockedByMe[u.ID] = true
	}

	blockedByPeer = make(map[uint]bool)
	for _, id := range blockers {
		blockedByPeer[id] = true
	}

	return blockedByMe, blockedByPeer, nil
}

