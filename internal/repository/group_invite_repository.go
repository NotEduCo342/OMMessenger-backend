package repository

import (
	"time"

	"github.com/noteduco342/OMMessenger-backend/internal/models"
	"gorm.io/gorm"
)

type GroupInviteRepository struct {
	db *gorm.DB
}

func NewGroupInviteRepository(db *gorm.DB) *GroupInviteRepository {
	return &GroupInviteRepository{db: db}
}

func (r *GroupInviteRepository) Create(link *models.GroupInviteLink) error {
	return r.db.Create(link).Error
}

func (r *GroupInviteRepository) FindByToken(token string) (*models.GroupInviteLink, error) {
	var link models.GroupInviteLink
	err := r.db.Where("token = ?", token).First(&link).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *GroupInviteRepository) IncrementUse(id uint) error {
	return r.db.Model(&models.GroupInviteLink{}).Where("id = ?", id).
		UpdateColumn("used_count", gorm.Expr("used_count + 1")).Error
}

func (r *GroupInviteRepository) Revoke(id uint, revokedAt time.Time) error {
	return r.db.Model(&models.GroupInviteLink{}).Where("id = ?", id).
		UpdateColumn("revoked_at", revokedAt).Error
}
