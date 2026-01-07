package models

import (
	"time"
)

type AppVersion struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	Platform          string    `gorm:"type:varchar(20);not null;index:idx_platform_active" json:"platform"`
	Version           string    `gorm:"type:varchar(20);not null" json:"version"`
	BuildNumber       int       `gorm:"not null;uniqueIndex:idx_platform_build" json:"build_number"`
	DownloadURL       string    `gorm:"type:text;not null" json:"download_url"`
	Changelog         string    `gorm:"type:text" json:"changelog,omitempty"`
	ForceUpdate       bool      `gorm:"default:false" json:"force_update"`
	MinSupportedBuild int       `gorm:"default:0" json:"min_supported_build"`
	IsActive          bool      `gorm:"default:true;index:idx_platform_active" json:"is_active"`
	ReleaseDate       time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"release_date"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (AppVersion) TableName() string {
	return "app_versions"
}

// ToResponse returns a client-friendly version response
func (v *AppVersion) ToResponse() map[string]interface{} {
	return map[string]interface{}{
		"version":             v.Version,
		"build_number":        v.BuildNumber,
		"download_url":        v.DownloadURL,
		"changelog":           v.Changelog,
		"force_update":        v.ForceUpdate,
		"min_supported_build": v.MinSupportedBuild,
		"release_date":        v.ReleaseDate.Format(time.RFC3339),
	}
}
