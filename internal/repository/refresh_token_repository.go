package repository

import (
	"time"

	"github.com/noteduco342/OMMessenger-backend/internal/models"
	"gorm.io/gorm"
)

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(token *models.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *RefreshTokenRepository) FindValidByHash(tokenHash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	if err := r.db.Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", tokenHash, time.Now()).First(&token).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *RefreshTokenRepository) RevokeByHash(tokenHash string) error {
	now := time.Now()
	return r.db.Model(&models.RefreshToken{}).
		Where("token_hash = ? AND revoked_at IS NULL", tokenHash).
		Update("revoked_at", &now).Error
}
