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
	db             *gorm.DB
	channelService *ChannelService
}

// NewVersionService creates a new version service instance
func NewVersionService() *VersionService {
	return &VersionService{
		db:             database.DB,
		channelService: NewChannelService(),
	}
}

// CreateVersion creates a new version
func (s *VersionService) CreateVersion(req *models.VersionRequest) (*models.Version, error) {
	// Check if version already exists for this app
	var existingVersion models.Version
	if err := s.db.Where("app_id = ? AND version = ?", req.AppID, req.Version).First(&existingVersion).Error; err == nil {
		return nil, fmt.Errorf("version %s already exists for this app", req.Version)
	}

	// Validate channel is enabled for this app
	if err := s.channelService.ValidateChannelForApp(req.AppID, req.Channel); err != nil {
		return nil, err
	}

	version := &models.Version{
		AppID:             req.AppID,
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

	// Check if channel has AutoPublish enabled for this app
	var appChannel models.ApplicationChannel
	if err := s.db.Where("app_id = ? AND channel_name = ?", req.AppID, req.Channel).First(&appChannel).Error; err == nil {
		if appChannel.AutoPublish {
			now := time.Now()
			version.IsPublished = true
			version.PublishTime = &now
			if err := s.db.Save(version).Error; err != nil {
				return nil, fmt.Errorf("failed to auto-publish version: %w", err)
			}
		}
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

	// Validate channel is enabled for this app
	if err := s.channelService.ValidateChannelForApp(version.AppID, req.Channel); err != nil {
		return nil, err
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

	// Validate that the associated channel is still active and enabled for this app
	if err := s.channelService.ValidateChannelForApp(version.AppID, version.Channel); err != nil {
		return nil, fmt.Errorf("failed to validate channel: %w", err)
	}

	// Get the global channel to check if it's active
	var channel models.Channel
	if err := s.db.Where("name = ?", version.Channel).First(&channel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("channel %s not found", version.Channel)
		}
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	if !channel.IsActive {
		return nil, fmt.Errorf("cannot publish version to inactive channel %s", version.Channel)
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

// GetLatestVersion gets the latest published version for a channel (deprecated, use GetLatestVersionForApp)
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

// GetLatestVersionForApp gets the latest published version for a specific app and channel
func (s *VersionService) GetLatestVersionForApp(appID, channel string) (*models.Version, error) {
	var version models.Version
	err := s.db.Where("app_id = ? AND channel = ? AND is_published = ?", appID, channel, true).
		Order("publish_time DESC").
		First(&version).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no published version found for app %s channel %s", appID, channel)
		}
		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}

	return &version, nil
}

// GetVersionsForApp gets published versions for a specific app with optional channel filter
func (s *VersionService) GetVersionsForApp(appID, channel string, limit int, publishedOnly bool) ([]*models.Version, error) {
	var versions []*models.Version

	query := s.db.Where("app_id = ?", appID)

	// Apply channel filter if provided
	if channel != "" {
		query = query.Where("channel = ?", channel)
	}

	// Apply published filter
	if publishedOnly {
		query = query.Where("is_published = ?", true)
	}

	// Apply limit (default 10, max 50)
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	// Order by publish_time (or created_at for unpublished) DESC and apply limit
	orderBy := "publish_time DESC"
	if !publishedOnly {
		orderBy = "COALESCE(publish_time, created_at) DESC"
	}

	if err := query.Order(orderBy).Limit(limit).Find(&versions).Error; err != nil {
		return nil, fmt.Errorf("failed to get versions for app: %w", err)
	}

	return versions, nil
}
