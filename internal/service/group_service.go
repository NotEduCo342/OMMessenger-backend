package service

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/noteduco342/OMMessenger-backend/internal/models"
	"github.com/noteduco342/OMMessenger-backend/internal/repository"
	"github.com/noteduco342/OMMessenger-backend/internal/validation"
	"gorm.io/gorm"
)

type GroupService struct {
	groupRepo          repository.GroupRepositoryInterface
	groupReadStateRepo repository.GroupReadStateRepositoryInterface
	userRepo           repository.UserRepositoryInterface
	inviteRepo         repository.GroupInviteRepositoryInterface
}

func NewGroupService(
	groupRepo repository.GroupRepositoryInterface,
	groupReadStateRepo repository.GroupReadStateRepositoryInterface,
	userRepo repository.UserRepositoryInterface,
	inviteRepo repository.GroupInviteRepositoryInterface,
) *GroupService {
	return &GroupService{
		groupRepo:          groupRepo,
		groupReadStateRepo: groupReadStateRepo,
		userRepo:           userRepo,
		inviteRepo:         inviteRepo,
	}
}

func (s *GroupService) CreateGroup(name, description string, creatorID uint) (*models.Group, error) {
	return s.CreateGroupWithVisibility(name, description, creatorID, false, "")
}

func (s *GroupService) CreateGroupWithVisibility(name, description string, creatorID uint, isPublic bool, handle string) (*models.Group, error) {
	group := &models.Group{
		Name:        name,
		Description: description,
		CreatorID:   creatorID,
		IsPublic:    isPublic,
	}

	if isPublic {
		if handle == "" {
			return nil, errors.New("handle is required for public groups")
		}
		if s.userRepo == nil {
			return nil, errors.New("user repository not configured")
		}
		normalized := validation.NormalizeHandle(handle)
		if !validation.ValidateHandle(normalized) {
			return nil, errors.New("invalid handle")
		}
		// Ensure handle not used by a user
		if _, err := s.userRepo.FindByUsername(normalized); err == nil {
			return nil, errors.New("handle already taken")
		}
		// Ensure handle not used by another group
		if _, err := s.groupRepo.FindByHandle(normalized); err == nil {
			return nil, errors.New("handle already taken")
		}
		group.Handle = &normalized
	}

	if err := s.groupRepo.Create(group); err != nil {
		return nil, err
	}

	// Add creator as admin
	if err := s.groupRepo.AddMember(group.ID, creatorID, models.RoleAdmin); err != nil {
		return nil, err
	}

	if s.groupReadStateRepo != nil {
		_ = s.groupReadStateRepo.EnsureForMember(group.ID, creatorID)
	}

	return s.groupRepo.FindByID(group.ID)
}

func (s *GroupService) UpdateGroup(groupID, requesterID uint, name, description string, isPublic bool, handle string) (*models.Group, error) {
	isAdmin, err := s.IsAdmin(groupID, requesterID)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, errors.New("forbidden")
	}

	group, err := s.groupRepo.FindByID(groupID)
	if err != nil {
		return nil, err
	}

	group.Name = name
	group.Description = description
	group.IsPublic = isPublic

	if isPublic {
		if handle == "" {
			return nil, errors.New("handle is required for public groups")
		}
		if s.userRepo == nil {
			return nil, errors.New("user repository not configured")
		}
		normalized := validation.NormalizeHandle(handle)
		if !validation.ValidateHandle(normalized) {
			return nil, errors.New("invalid handle")
		}
		// Ensure handle not used by a user
		if _, err := s.userRepo.FindByUsername(normalized); err == nil {
			return nil, errors.New("handle already taken")
		}
		// Ensure handle not used by another group (excluding self)
		if existing, err := s.groupRepo.FindByHandle(normalized); err == nil && existing.ID != groupID {
			return nil, errors.New("handle already taken")
		}
		group.Handle = &normalized
	} else {
		group.Handle = nil
	}

	if err := s.groupRepo.Update(group); err != nil {
		return nil, err
	}

	return s.groupRepo.FindByID(group.ID)
}

func (s *GroupService) IsHandleAvailable(handle string, excludeGroupID uint) (bool, error) {
	normalized := validation.NormalizeHandle(handle)
	if !validation.ValidateHandle(normalized) {
		return false, nil
	}

	// 1. Ensure handle not used by a user
	if s.userRepo != nil {
		if _, err := s.userRepo.FindByUsername(normalized); err == nil {
			return false, nil
		}
	}

	// 2. Ensure handle not used by another group (excluding self)
	if s.groupRepo != nil {
		if existing, err := s.groupRepo.FindByHandle(normalized); err == nil {
			if excludeGroupID == 0 || existing.ID != excludeGroupID {
				return false, nil
			}
		}
	}

	return true, nil
}

func (s *GroupService) AddMemberDirectly(groupID, requesterID, userID uint) error {
	isAdmin, err := s.IsAdmin(groupID, requesterID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errors.New("forbidden")
	}

	isMember, err := s.groupRepo.IsMember(groupID, userID)
	if err != nil {
		return err
	}
	if isMember {
		return errors.New("user is already a member of this group")
	}

	if err := s.groupRepo.AddMember(groupID, userID, models.RoleMember); err != nil {
		return err
	}
	if s.groupReadStateRepo != nil {
		_ = s.groupReadStateRepo.EnsureForMember(groupID, userID)
	}
	return nil
}

func (s *GroupService) RemoveMemberDirectly(groupID, requesterID, userID uint) error {
	isAdmin, err := s.IsAdmin(groupID, requesterID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errors.New("forbidden")
	}

	// Optionally prevent removing the last admin
	if requesterID == userID {
		// Can't remove yourself via this endpoint, use LeaveGroup
		return errors.New("cannot remove yourself using this endpoint")
	}

	if err := s.groupRepo.RemoveMember(groupID, userID); err != nil {
		return err
	}
	if s.groupReadStateRepo != nil {
		_ = s.groupReadStateRepo.DeleteForMember(groupID, userID)
	}
	return nil
}

func (s *GroupService) JoinGroup(groupID, userID uint) error {
	group, err := s.groupRepo.FindByID(groupID)
	if err != nil {
		return err
	}
	if !group.IsPublic {
		return errors.New("group is private")
	}

	// Check if already a member
	isMember, err := s.groupRepo.IsMember(groupID, userID)
	if err != nil {
		return err
	}
	if isMember {
		return errors.New("user is already a member of this group")
	}

	if err := s.groupRepo.AddMember(groupID, userID, models.RoleMember); err != nil {
		return err
	}
	if s.groupReadStateRepo != nil {
		_ = s.groupReadStateRepo.EnsureForMember(groupID, userID)
	}
	return nil
}

func (s *GroupService) JoinGroupByHandle(handle string, userID uint) (*models.Group, error) {
	if handle == "" {
		return nil, errors.New("handle is required")
	}
	group, err := s.groupRepo.FindByHandle(handle)
	if err != nil {
		return nil, err
	}
	if !group.IsPublic {
		return nil, errors.New("group is private")
	}
	if err := s.JoinGroup(group.ID, userID); err != nil {
		return nil, err
	}
	return group, nil
}

func (s *GroupService) EnsureReadState(groupID, userID uint) {
	if s.groupReadStateRepo == nil {
		return
	}
	_ = s.groupReadStateRepo.EnsureForMember(groupID, userID)
}

func (s *GroupService) LeaveGroup(groupID, userID uint) error {
	if err := s.groupRepo.RemoveMember(groupID, userID); err != nil {
		return err
	}
	if s.groupReadStateRepo != nil {
		_ = s.groupReadStateRepo.DeleteForMember(groupID, userID)
	}
	return nil
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

func (s *GroupService) DeleteGroup(groupID uint) error {
	return s.groupRepo.Delete(groupID)
}

func (s *GroupService) GetPublicGroupByHandle(handle string) (*models.Group, error) {
	group, err := s.groupRepo.FindByHandle(handle)
	if err != nil {
		return nil, err
	}
	if !group.IsPublic {
		return nil, errors.New("group is private")
	}
	return group, nil
}

func (s *GroupService) SearchPublicGroups(query string, limit int) ([]models.Group, error) {
	return s.groupRepo.SearchPublicGroups(query, limit)
}

func (s *GroupService) IsMember(groupID, userID uint) (bool, error) {
	return s.groupRepo.IsMember(groupID, userID)
}

func (s *GroupService) IsAdmin(groupID, userID uint) (bool, error) {
	role, err := s.groupRepo.GetMemberRole(groupID, userID)
	if err != nil {
		return false, err
	}
	return role == models.RoleAdmin, nil
}

func (s *GroupService) UpsertReadStateMonotonic(groupID, userID, lastReadMessageID uint) error {
	if s.groupReadStateRepo == nil {
		return nil
	}
	return s.groupReadStateRepo.UpsertMonotonic(groupID, userID, lastReadMessageID)
}

func (s *GroupService) GetReadState(groupID, userID uint) (*models.GroupReadState, error) {
	if s.groupReadStateRepo == nil {
		return &models.GroupReadState{GroupID: groupID, UserID: userID, LastReadMessageID: 0}, nil
	}
	state, err := s.groupReadStateRepo.Get(groupID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &models.GroupReadState{GroupID: groupID, UserID: userID, LastReadMessageID: 0}, nil
		}
		return nil, err
	}
	return state, nil
}

func (s *GroupService) ListReadStates(groupID uint) ([]models.GroupReadState, error) {
	if s.groupReadStateRepo == nil {
		return []models.GroupReadState{}, nil
	}
	return s.groupReadStateRepo.ListByGroup(groupID)
}

func (s *GroupService) CreateInviteLink(groupID, creatorID uint, singleUse bool, expiresAt *time.Time) (*models.GroupInviteLink, error) {
	if s.inviteRepo == nil {
		return nil, errors.New("invite repository not configured")
	}
	// Only admins can create invite links
	isAdmin, err := s.IsAdmin(groupID, creatorID)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, errors.New("forbidden")
	}

	maxUses := (*int)(nil)
	if singleUse {
		v := 1
		maxUses = &v
	}

	token, err := generateInviteToken()
	if err != nil {
		return nil, err
	}

	link := &models.GroupInviteLink{
		GroupID:   groupID,
		Token:     token,
		CreatedBy: creatorID,
		ExpiresAt: expiresAt,
		MaxUses:   maxUses,
		UsedCount: 0,
	}

	if err := s.inviteRepo.Create(link); err != nil {
		return nil, err
	}
	return link, nil
}

func (s *GroupService) JoinGroupByInvite(token string, userID uint) (*models.Group, error) {
	if s.inviteRepo == nil {
		return nil, errors.New("invite repository not configured")
	}
	link, err := s.inviteRepo.FindByToken(token)
	if err != nil {
		return nil, err
	}
	if link.RevokedAt != nil {
		return nil, errors.New("invite link revoked")
	}
	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		return nil, errors.New("invite link expired")
	}
	if link.MaxUses != nil && link.UsedCount >= *link.MaxUses {
		return nil, errors.New("invite link exhausted")
	}

	group, err := s.groupRepo.FindByID(link.GroupID)
	if err != nil {
		return nil, err
	}

	// Check if already a member
	isMember, err := s.groupRepo.IsMember(group.ID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		if err := s.groupRepo.AddMember(group.ID, userID, models.RoleMember); err != nil {
			return nil, err
		}
		if s.groupReadStateRepo != nil {
			_ = s.groupReadStateRepo.EnsureForMember(group.ID, userID)
		}
		if err := s.inviteRepo.IncrementUse(link.ID); err != nil {
			return nil, err
		}
	}
	return group, nil
}

func (s *GroupService) GetInvitePreview(token string) (*models.GroupInviteLink, *models.Group, error) {
	if s.inviteRepo == nil {
		return nil, nil, errors.New("invite repository not configured")
	}
	link, err := s.inviteRepo.FindByToken(token)
	if err != nil {
		return nil, nil, err
	}
	if link.RevokedAt != nil {
		return nil, nil, errors.New("invite link revoked")
	}
	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		return nil, nil, errors.New("invite link expired")
	}
	if link.MaxUses != nil && link.UsedCount >= *link.MaxUses {
		return nil, nil, errors.New("invite link exhausted")
	}

	group, err := s.groupRepo.FindByID(link.GroupID)
	if err != nil {
		return nil, nil, err
	}
	return link, group, nil
}

func generateInviteToken() (string, error) {
	b := make([]byte, 18)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
