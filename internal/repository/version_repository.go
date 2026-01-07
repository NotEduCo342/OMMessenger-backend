package repository

import (
	"github.com/noteduco342/OMMessenger-backend/internal/models"
	"gorm.io/gorm"
)

type VersionRepository struct {
	db *gorm.DB
}

func NewVersionRepository(db *gorm.DB) *VersionRepository {
	return &VersionRepository{db: db}
}

// GetActiveVersion returns the currently active version for a platform
func (r *VersionRepository) GetActiveVersion(platform string) (*models.AppVersion, error) {
	var version models.AppVersion
	err := r.db.Where("platform = ? AND is_active = ?", platform, true).
		First(&version).Error

	if err != nil {
		return nil, err
	}

	return &version, nil
}

// GetVersionByBuildNumber retrieves a specific version by platform and build number
func (r *VersionRepository) GetVersionByBuildNumber(platform string, buildNumber int) (*models.AppVersion, error) {
	var version models.AppVersion
	err := r.db.Where("platform = ? AND build_number = ?", platform, buildNumber).
		First(&version).Error

	if err != nil {
		return nil, err
	}

	return &version, nil
}

// GetAllVersions returns all versions for a platform (for admin purposes)
func (r *VersionRepository) GetAllVersions(platform string) ([]models.AppVersion, error) {
	var versions []models.AppVersion
	err := r.db.Where("platform = ?", platform).
		Order("build_number DESC").
		Find(&versions).Error

	return versions, err
}

// CreateVersion creates a new version entry
func (r *VersionRepository) CreateVersion(version *models.AppVersion) error {
	return r.db.Create(version).Error
}

// SetActiveVersion sets a version as active and deactivates all others for the platform
func (r *VersionRepository) SetActiveVersion(platform string, buildNumber int) error {
	// Use transaction to ensure atomicity
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Deactivate all versions for this platform
		if err := tx.Model(&models.AppVersion{}).
			Where("platform = ?", platform).
			Update("is_active", false).Error; err != nil {
			return err
		}

		// Activate the specified version
		if err := tx.Model(&models.AppVersion{}).
			Where("platform = ? AND build_number = ?", platform, buildNumber).
			Update("is_active", true).Error; err != nil {
			return err
		}

		return nil
	})
}

// UpdateVersion updates an existing version
func (r *VersionRepository) UpdateVersion(version *models.AppVersion) error {
	return r.db.Save(version).Error
}
