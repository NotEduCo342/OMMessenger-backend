package service

import (
	"errors"

	"github.com/noteduco342/OMMessenger-backend/internal/models"
	"github.com/noteduco342/OMMessenger-backend/internal/repository"
)

type GroupService struct {
	groupRepo repository.GroupRepositoryInterface
}

func NewGroupService(groupRepo repository.GroupRepositoryInterface) *GroupService {
	return &GroupService{groupRepo: groupRepo}
}

func (s *GroupService) CreateGroup(name, description string, creatorID uint) (*models.Group, error) {
	group := &models.Group{
		Name:        name,
		Description: description,
		CreatorID:   creatorID,
	}

	if err := s.groupRepo.Create(group); err != nil {
		return nil, err
	}

	// Add creator as admin
	if err := s.groupRepo.AddMember(group.ID, creatorID, models.RoleAdmin); err != nil {
		return nil, err
	}

	return s.groupRepo.FindByID(group.ID)
}

func (s *GroupService) JoinGroup(groupID, userID uint) error {
	// Check if already a member
	isMember, err := s.groupRepo.IsMember(groupID, userID)
	if err != nil {
		return err
	}
	if isMember {
		return errors.New("user is already a member of this group")
	}

	return s.groupRepo.AddMember(groupID, userID, models.RoleMember)
}

func (s *GroupService) LeaveGroup(groupID, userID uint) error {
	return s.groupRepo.RemoveMember(groupID, userID)
}

func (s *GroupService) GetGroupMembers(groupID uint) ([]models.User, error) {
	return s.groupRepo.GetMembers(groupID)
}

func (s *GroupService) GetUserGroups(userID uint) ([]models.Group, error) {
	return s.groupRepo.GetUserGroups(userID)
}

func (s *GroupService) GetGroup(groupID uint) (*models.Group, error) {
	return s.groupRepo.FindByID(groupID)
}
