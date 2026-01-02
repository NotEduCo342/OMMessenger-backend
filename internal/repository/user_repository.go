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
	err := r.db.Where("username = ?", username).First(&user).Error
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
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("is_online", isOnline).Error
}

func (r *UserRepository) SearchUsers(query string, limit int) ([]models.User, error) {
	var users []models.User

	// Search by username or full name (case insensitive)
	err := r.db.Where("LOWER(username) LIKE ? OR LOWER(full_name) LIKE ?", "%"+query+"%", "%"+query+"%").
		Limit(limit).
		Find(&users).Error

	return users, err
}
