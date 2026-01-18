package repository

import (
	"github.com/noteduco342/OMMessenger-backend/internal/models"
	"gorm.io/gorm"
)

type GroupReadStateRepository struct {
	db *gorm.DB
}

func NewGroupReadStateRepository(db *gorm.DB) *GroupReadStateRepository {
	return &GroupReadStateRepository{db: db}
}

func (r *GroupReadStateRepository) EnsureForMember(groupID, userID uint) error {
	return r.db.Exec(`
		INSERT INTO group_read_states (group_id, user_id, last_read_message_id, created_at, updated_at)
		VALUES (?, ?, 0, NOW(), NOW())
		ON CONFLICT (group_id, user_id) DO NOTHING
	`, groupID, userID).Error
}

func (r *GroupReadStateRepository) DeleteForMember(groupID, userID uint) error {
	return r.db.Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&models.GroupReadState{}).Error
}

func (r *GroupReadStateRepository) UpsertMonotonic(groupID, userID uint, lastReadMessageID uint) error {
	return r.db.Exec(`
		INSERT INTO group_read_states (group_id, user_id, last_read_message_id, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())
		ON CONFLICT (group_id, user_id) DO UPDATE
		SET last_read_message_id = GREATEST(group_read_states.last_read_message_id, EXCLUDED.last_read_message_id),
			updated_at = NOW()
	`, groupID, userID, lastReadMessageID).Error
}

func (r *GroupReadStateRepository) Get(groupID, userID uint) (*models.GroupReadState, error) {
	var state models.GroupReadState
	err := r.db.Where("group_id = ? AND user_id = ?", groupID, userID).First(&state).Error
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func (r *GroupReadStateRepository) ListByGroup(groupID uint) ([]models.GroupReadState, error) {
	var states []models.GroupReadState
	err := r.db.Where("group_id = ?", groupID).Find(&states).Error
	return states, err
}
