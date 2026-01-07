package service

import (
	"fmt"

	"github.com/noteduco342/OMMessenger-backend/internal/models"
	"github.com/noteduco342/OMMessenger-backend/internal/repository"
)

type VersionService struct {
	versionRepo *repository.VersionRepository
}

func NewVersionService(versionRepo *repository.VersionRepository) *VersionService {
	return &VersionService{
		versionRepo: versionRepo,
	}
}

// GetLatestVersion returns the active version for a platform
func (s *VersionService) GetLatestVersion(platform string) (*models.AppVersion, error) {
	// Validate platform
	if platform != "android" && platform != "ios" && platform != "web" {
		return nil, fmt.Errorf("invalid platform: %s", platform)
	}

	version, err := s.versionRepo.GetActiveVersion(platform)
	if err != nil {
		return nil, fmt.Errorf("failed to get active version: %w", err)
	}

	return version, nil
}

// CheckUpdateRequired determines if an update is needed based on build number
func (s *VersionService) CheckUpdateRequired(platform string, currentBuild int) (bool, *models.AppVersion, error) {
	latestVersion, err := s.GetLatestVersion(platform)
	if err != nil {
		return false, nil, err
	}

	// Update needed if current build is lower than latest
	needsUpdate := currentBuild < latestVersion.BuildNumber

	return needsUpdate, latestVersion, nil
}

// IsForceUpdateRequired checks if the current build MUST update
func (s *VersionService) IsForceUpdateRequired(platform string, currentBuild int) (bool, error) {
	latestVersion, err := s.GetLatestVersion(platform)
	if err != nil {
		return false, err
	}

	// Force update if:
	// 1. Current build is below minimum supported build
	// 2. OR force_update flag is set for latest version
	if currentBuild < latestVersion.MinSupportedBuild {
		return true, nil
	}

	if latestVersion.ForceUpdate && currentBuild < latestVersion.BuildNumber {
		return true, nil
	}

	return false, nil
}
