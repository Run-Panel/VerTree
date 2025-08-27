package services

import (
	"fmt"
	"time"

	"github.com/Run-Panel/VerTree/internal/database"
	"github.com/Run-Panel/VerTree/internal/models"
	"gorm.io/gorm"
)

// VersionService handles version-related business logic
type VersionService struct {
	db *gorm.DB
}

// NewVersionService creates a new version service instance
func NewVersionService() *VersionService {
	return &VersionService{
		db: database.DB,
	}
}

// CreateVersion creates a new version
func (s *VersionService) CreateVersion(req *models.VersionRequest) (*models.Version, error) {
	// Check if version already exists
	var existingVersion models.Version
	if err := s.db.Where("version = ?", req.Version).First(&existingVersion).Error; err == nil {
		return nil, fmt.Errorf("version %s already exists", req.Version)
	}

	// Validate channel exists
	var channel models.Channel
	if err := s.db.Where("name = ? AND is_active = ?", req.Channel, true).First(&channel).Error; err != nil {
		return nil, fmt.Errorf("invalid or inactive channel: %s", req.Channel)
	}

	version := &models.Version{
		Version:           req.Version,
		Channel:           req.Channel,
		Title:             req.Title,
		Description:       req.Description,
		ReleaseNotes:      req.ReleaseNotes,
		BreakingChanges:   req.BreakingChanges,
		MinUpgradeVersion: req.MinUpgradeVersion,
		FileURL:           req.FileURL,
		FileSize:          req.FileSize,
		FileChecksum:      req.FileChecksum,
		IsForced:          req.IsForced,
		IsPublished:       false,
	}

	if err := s.db.Create(version).Error; err != nil {
		return nil, fmt.Errorf("failed to create version: %w", err)
	}

	return version, nil
}

// GetVersionByID retrieves a version by ID
func (s *VersionService) GetVersionByID(id uint) (*models.Version, error) {
	var version models.Version
	if err := s.db.First(&version, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("version not found")
		}
		return nil, fmt.Errorf("failed to get version: %w", err)
	}
	return &version, nil
}

// UpdateVersion updates an existing version
func (s *VersionService) UpdateVersion(id uint, req *models.VersionRequest) (*models.Version, error) {
	version, err := s.GetVersionByID(id)
	if err != nil {
		return nil, err
	}

	// Don't allow updating published versions
	if version.IsPublished {
		return nil, fmt.Errorf("cannot update published version")
	}

	// Check if new version number conflicts
	if req.Version != version.Version {
		var existingVersion models.Version
		if err := s.db.Where("version = ? AND id != ?", req.Version, id).First(&existingVersion).Error; err == nil {
			return nil, fmt.Errorf("version %s already exists", req.Version)
		}
	}

	// Validate channel exists
	var channel models.Channel
	if err := s.db.Where("name = ? AND is_active = ?", req.Channel, true).First(&channel).Error; err != nil {
		return nil, fmt.Errorf("invalid or inactive channel: %s", req.Channel)
	}

	// Update fields
	version.Version = req.Version
	version.Channel = req.Channel
	version.Title = req.Title
	version.Description = req.Description
	version.ReleaseNotes = req.ReleaseNotes
	version.BreakingChanges = req.BreakingChanges
	version.MinUpgradeVersion = req.MinUpgradeVersion
	version.FileURL = req.FileURL
	version.FileSize = req.FileSize
	version.FileChecksum = req.FileChecksum
	version.IsForced = req.IsForced

	if err := s.db.Save(version).Error; err != nil {
		return nil, fmt.Errorf("failed to update version: %w", err)
	}

	return version, nil
}

// PublishVersion publishes a version
func (s *VersionService) PublishVersion(id uint) (*models.Version, error) {
	version, err := s.GetVersionByID(id)
	if err != nil {
		return nil, err
	}

	if version.IsPublished {
		return nil, fmt.Errorf("version is already published")
	}

	now := time.Now()
	version.IsPublished = true
	version.PublishTime = &now

	if err := s.db.Save(version).Error; err != nil {
		return nil, fmt.Errorf("failed to publish version: %w", err)
	}

	return version, nil
}

// UnpublishVersion unpublishes a version
func (s *VersionService) UnpublishVersion(id uint) (*models.Version, error) {
	version, err := s.GetVersionByID(id)
	if err != nil {
		return nil, err
	}

	if !version.IsPublished {
		return nil, fmt.Errorf("version is not published")
	}

	version.IsPublished = false
	version.PublishTime = nil

	if err := s.db.Save(version).Error; err != nil {
		return nil, fmt.Errorf("failed to unpublish version: %w", err)
	}

	return version, nil
}

// DeleteVersion deletes a version
func (s *VersionService) DeleteVersion(id uint) error {
	version, err := s.GetVersionByID(id)
	if err != nil {
		return err
	}

	if version.IsPublished {
		return fmt.Errorf("cannot delete published version")
	}

	if err := s.db.Delete(version).Error; err != nil {
		return fmt.Errorf("failed to delete version: %w", err)
	}

	return nil
}

// ListVersions lists versions with pagination and filters
func (s *VersionService) ListVersions(channel string, page, limit int) ([]*models.Version, int64, error) {
	var versions []*models.Version
	var total int64

	query := s.db.Model(&models.Version{})

	// Apply channel filter if provided
	if channel != "" {
		query = query.Where("channel = ?", channel)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count versions: %w", err)
	}

	// Apply pagination and ordering
	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&versions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list versions: %w", err)
	}

	return versions, total, nil
}

// GetLatestVersion gets the latest published version for a channel
func (s *VersionService) GetLatestVersion(channel string) (*models.Version, error) {
	var version models.Version
	err := s.db.Where("channel = ? AND is_published = ?", channel, true).
		Order("publish_time DESC").
		First(&version).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no published version found for channel %s", channel)
		}
		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}

	return &version, nil
}
