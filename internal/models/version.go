package models

import (
	"time"

	"gorm.io/gorm"
)

// Version represents a software version in the database
type Version struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	AppID             string         `json:"app_id" gorm:"size:32;uniqueIndex:idx_app_version" validate:"required"`
	Version           string         `json:"version" gorm:"not null;size:50;uniqueIndex:idx_app_version" validate:"required"`
	Channel           string         `json:"channel" gorm:"not null;size:20;default:stable" validate:"required"`
	Title             string         `json:"title" gorm:"not null;size:200" validate:"required"`
	Description       string         `json:"description" gorm:"type:text"`
	ReleaseNotes      string         `json:"release_notes" gorm:"type:text"`
	BreakingChanges   string         `json:"breaking_changes" gorm:"type:text"`
	MinUpgradeVersion string         `json:"min_upgrade_version" gorm:"size:50"`
	FileURL           string         `json:"file_url" gorm:"not null;size:500" validate:"required,url"`
	FileSize          int64          `json:"file_size" gorm:"not null" validate:"required,min=1"`
	FileChecksum      string         `json:"file_checksum" gorm:"not null;size:128" validate:"required"`
	IsPublished       bool           `json:"is_published" gorm:"default:false"`
	IsForced          bool           `json:"is_forced" gorm:"default:false"`
	PublishTime       *time.Time     `json:"publish_time"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Application Application `json:"application,omitempty" gorm:"foreignKey:AppID;references:AppID"`
}

// TableName returns the table name for Version model
func (Version) TableName() string {
	return "versions"
}

// VersionRequest represents the request payload for creating/updating versions
type VersionRequest struct {
	AppID             string `json:"app_id" validate:"required"`
	Version           string `json:"version" validate:"required"`
	Channel           string `json:"channel" validate:"required"`
	Title             string `json:"title" validate:"required"`
	Description       string `json:"description"`
	ReleaseNotes      string `json:"release_notes"`
	BreakingChanges   string `json:"breaking_changes"`
	MinUpgradeVersion string `json:"min_upgrade_version"`
	FileURL           string `json:"file_url" validate:"required,url"`
	FileSize          int64  `json:"file_size" validate:"required,min=1"`
	FileChecksum      string `json:"file_checksum" validate:"required"`
	IsForced          bool   `json:"is_forced"`
}

// VersionResponse represents the response payload for version queries
type VersionResponse struct {
	ID                uint       `json:"id"`
	AppID             string     `json:"app_id"`
	Version           string     `json:"version"`
	Channel           string     `json:"channel"`
	Title             string     `json:"title"`
	Description       string     `json:"description"`
	ReleaseNotes      string     `json:"release_notes"`
	BreakingChanges   string     `json:"breaking_changes"`
	MinUpgradeVersion string     `json:"min_upgrade_version"`
	FileURL           string     `json:"file_url"`
	FileSize          int64      `json:"file_size"`
	FileChecksum      string     `json:"file_checksum"`
	IsPublished       bool       `json:"is_published"`
	IsForced          bool       `json:"is_forced"`
	PublishTime       *time.Time `json:"publish_time"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// ToResponse converts Version model to VersionResponse
func (v *Version) ToResponse() *VersionResponse {
	return &VersionResponse{
		ID:                v.ID,
		AppID:             v.AppID,
		Version:           v.Version,
		Channel:           v.Channel,
		Title:             v.Title,
		Description:       v.Description,
		ReleaseNotes:      v.ReleaseNotes,
		BreakingChanges:   v.BreakingChanges,
		MinUpgradeVersion: v.MinUpgradeVersion,
		FileURL:           v.FileURL,
		FileSize:          v.FileSize,
		FileChecksum:      v.FileChecksum,
		IsPublished:       v.IsPublished,
		IsForced:          v.IsForced,
		PublishTime:       v.PublishTime,
		CreatedAt:         v.CreatedAt,
		UpdatedAt:         v.UpdatedAt,
	}
}
