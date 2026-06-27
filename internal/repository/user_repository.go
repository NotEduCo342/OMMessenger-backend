package repository

import (
	"github.com/noteduco342/OMMessenger-backend/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("LOWER(username) = LOWER(?)", username).First(&user).Error
	return &user, err
}

func (r *UserRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	return &user, err
}

func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) UpdateOnlineStatus(userID uint, isOnline bool) error {
	updates := map[string]interface{}{
		"is_online": isOnline,
	}

	if !isOnline {
		updates["last_seen"] = gorm.Expr("NOW()")
	}

	return r.db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error
}

func (r *UserRepository) SearchUsers(query string, limit int) ([]models.User, error) {
	var users []models.User

	// Search by username or full name (case insensitive)
	err := r.db.Where("LOWER(username) LIKE ? OR LOWER(full_name) LIKE ?", query+"%", query+"%").
		Limit(limit).
		Find(&users).Error

	return users, err
}

func (r *UserRepository) BlockUser(blockerID, blockedID uint) error {
	block := models.Block{
		BlockerID: blockerID,
		BlockedID: blockedID,
	}
	return r.db.Create(&block).Error
}

func (r *UserRepository) UnblockUser(blockerID, blockedID uint) error {
	return r.db.Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).Delete(&models.Block{}).Error
}

func (r *UserRepository) IsBlocked(userID1, userID2 uint) (bool, error) {
	var dummy int
	// Optimization: Use Select("1").Limit(1).Scan() instead of Count() to fail fast when checking for existence
	err := r.db.Model(&models.Block{}).
		Select("1").
		Where("(blocker_id = ? AND blocked_id = ?) OR (blocker_id = ? AND blocked_id = ?)", userID1, userID2, userID2, userID1).
		Limit(1).
		Scan(&dummy).Error
	return dummy > 0, err
}

func (r *UserRepository) IsBlocker(blockerID, blockedID uint) (bool, error) {
	var dummy int
	// Optimization: Use Select("1").Limit(1).Scan() instead of Count() to fail fast when checking for existence
	err := r.db.Model(&models.Block{}).
		Select("1").
		Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).
		Limit(1).
		Scan(&dummy).Error
	return dummy > 0, err
}

func (r *UserRepository) GetBlockedUsers(userID uint) ([]models.User, error) {
	var users []models.User
	err := r.db.Table("users").
		Joins("JOIN blocks ON blocks.blocked_id = users.id").
		Where("blocks.blocker_id = ?", userID).
		Find(&users).Error
	return users, err
}

func (r *UserRepository) GetBlockRelationshipsForUser(userID uint) ([]uint, error) {
	var ids []uint
	err := r.db.Model(&models.Block{}).
		Where("blocker_id = ? OR blocked_id = ?", userID, userID).
		Select("CASE WHEN blocker_id = ? THEN blocked_id ELSE blocker_id END", userID, userID).
		Find(&ids).Error
	return ids, err
}

func (r *UserRepository) GetBlockerIDs(userID uint) ([]uint, error) {
	var ids []uint
	err := r.db.Model(&models.Block{}).
		Where("blocked_id = ?", userID).
		Pluck("blocker_id", &ids).Error
	return ids, err
}
