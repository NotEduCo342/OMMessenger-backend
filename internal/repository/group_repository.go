package repository

import (
	"github.com/noteduco342/OMMessenger-backend/internal/models"
	"gorm.io/gorm"
)

type GroupRepository struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

func (r *GroupRepository) Create(group *models.Group) error {
	return r.db.Create(group).Error
}

func (r *GroupRepository) FindByID(id uint) (*models.Group, error) {
	var group models.Group
	if err := r.db.Preload("Members").Preload("Creator").First(&group, id).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *GroupRepository) FindByHandle(handle string) (*models.Group, error) {
	var group models.Group
	err := r.db.Where("LOWER(handle) = LOWER(?)", handle).
		Preload("Creator").
		First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *GroupRepository) AddMember(groupID, userID uint, role models.GroupRole) error {
	member := models.GroupMember{
		GroupID: groupID,
		UserID:  userID,
		Role:    role,
	}
	return r.db.Create(&member).Error
}

func (r *GroupRepository) RemoveMember(groupID, userID uint) error {
	return r.db.Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&models.GroupMember{}).Error
}

func (r *GroupRepository) GetMembers(groupID uint) ([]models.User, error) {
	var members []models.User
	err := r.db.Joins("JOIN group_members ON group_members.user_id = users.id").
		Where("group_members.group_id = ?", groupID).
		Find(&members).Error
	return members, err
}

func (r *GroupRepository) IsMember(groupID, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *GroupRepository) GetMemberRole(groupID, userID uint) (models.GroupRole, error) {
	var member models.GroupMember
	if err := r.db.Where("group_id = ? AND user_id = ?", groupID, userID).First(&member).Error; err != nil {
		return "", err
	}
	return member.Role, nil
}

func (r *GroupRepository) GetUserGroups(userID uint) ([]models.Group, error) {
	var groups []models.Group
	err := r.db.Joins("JOIN group_members ON group_members.group_id = groups.id").
		Where("group_members.user_id = ?", userID).
		Preload("Creator").
		Find(&groups).Error
	return groups, err
}

func (r *GroupRepository) SearchPublicGroups(query string, limit int) ([]models.Group, error) {
	var groups []models.Group
	q := "%" + query + "%"
	err := r.db.Where("is_public = true AND (LOWER(handle) LIKE LOWER(?) OR LOWER(name) LIKE LOWER(?))", q, q).
		Limit(limit).
		Preload("Creator").
		Find(&groups).Error
	return groups, err
}
