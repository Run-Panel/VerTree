package services

import (
	"fmt"

	"github.com/Run-Panel/VerTree/internal/database"
	"github.com/Run-Panel/VerTree/internal/models"
	"gorm.io/gorm"
)

// ChannelService handles channel-related business logic
type ChannelService struct {
	db *gorm.DB
}

// NewChannelService creates a new channel service instance
func NewChannelService() *ChannelService {
	return &ChannelService{
		db: database.DB,
	}
}

// GetAllChannels retrieves all global channels
func (s *ChannelService) GetAllChannels() ([]*models.Channel, error) {
	var channels []*models.Channel
	if err := s.db.Order("sort_order, name").Find(&channels).Error; err != nil {
		return nil, fmt.Errorf("failed to get channels: %w", err)
	}
	return channels, nil
}

// GetChannelByID retrieves a global channel by ID
func (s *ChannelService) GetChannelByID(id uint) (*models.Channel, error) {
	var channel models.Channel
	if err := s.db.First(&channel, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("channel not found")
		}
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}
	return &channel, nil
}

// GetChannelByName retrieves a global channel by name
func (s *ChannelService) GetChannelByName(name string) (*models.Channel, error) {
	var channel models.Channel
	if err := s.db.Where("name = ?", name).First(&channel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("channel not found")
		}
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}
	return &channel, nil
}

// GetChannelsByApp retrieves all enabled channels for a specific app
func (s *ChannelService) GetChannelsByApp(appID string) ([]*models.ApplicationChannelResponse, error) {
	var appChannels []models.ApplicationChannel

	// Get all enabled channels for this app with channel details
	if err := s.db.Preload("Channel").
		Where("app_id = ? AND is_enabled = ?", appID, true).
		Order("channel_name").
		Find(&appChannels).Error; err != nil {
		return nil, fmt.Errorf("failed to get channels for app %s: %w", appID, err)
	}

	// Convert to response format
	var responses []*models.ApplicationChannelResponse
	for _, appChannel := range appChannels {
		responses = append(responses, appChannel.ToResponse())
	}

	return responses, nil
}

// GetAllChannelsForApp retrieves all channels (enabled and disabled) for a specific app
func (s *ChannelService) GetAllChannelsForApp(appID string) ([]*models.ApplicationChannelResponse, error) {
	var appChannels []models.ApplicationChannel

	// Get all channels for this app with channel details
	if err := s.db.Preload("Channel").
		Where("app_id = ?", appID).
		Order("channel_name").
		Find(&appChannels).Error; err != nil {
		return nil, fmt.Errorf("failed to get all channels for app %s: %w", appID, err)
	}

	// Convert to response format
	var responses []*models.ApplicationChannelResponse
	for _, appChannel := range appChannels {
		responses = append(responses, appChannel.ToResponse())
	}

	return responses, nil
}

// CreateChannel creates a new global channel (admin only)
func (s *ChannelService) CreateChannel(req *models.ChannelRequest) (*models.Channel, error) {
	// Check if channel name already exists globally
	var existingChannel models.Channel
	if err := s.db.Where("name = ?", req.Name).First(&existingChannel).Error; err == nil {
		return nil, fmt.Errorf("channel name %s already exists", req.Name)
	}

	channel := &models.Channel{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		IsActive:    req.IsActive,
		SortOrder:   req.SortOrder,
	}

	if err := s.db.Create(channel).Error; err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	return channel, nil
}

// UpdateChannel updates an existing global channel (admin only)
func (s *ChannelService) UpdateChannel(id uint, req *models.ChannelRequest) (*models.Channel, error) {
	channel, err := s.GetChannelByID(id)
	if err != nil {
		return nil, err
	}

	// Check if new name conflicts globally (if changed)
	if req.Name != channel.Name {
		var existingChannel models.Channel
		if err := s.db.Where("name = ? AND id != ?", req.Name, id).First(&existingChannel).Error; err == nil {
			return nil, fmt.Errorf("channel name %s already exists", req.Name)
		}
	}

	// Update fields
	channel.Name = req.Name
	channel.DisplayName = req.DisplayName
	channel.Description = req.Description
	channel.IsActive = req.IsActive
	channel.SortOrder = req.SortOrder

	if err := s.db.Save(channel).Error; err != nil {
		return nil, fmt.Errorf("failed to update channel: %w", err)
	}

	return channel, nil
}

// DeleteChannel deletes a global channel (admin only)
func (s *ChannelService) DeleteChannel(id uint) error {
	channel, err := s.GetChannelByID(id)
	if err != nil {
		return err
	}

	// Check if channel has any versions
	var versionCount int64
	if err := s.db.Model(&models.Version{}).Where("channel = ?", channel.Name).Count(&versionCount).Error; err != nil {
		return fmt.Errorf("failed to check version count: %w", err)
	}

	if versionCount > 0 {
		return fmt.Errorf("cannot delete channel with existing versions")
	}

	// Delete all application-channel relationships first
	if err := s.db.Where("channel_name = ?", channel.Name).Delete(&models.ApplicationChannel{}).Error; err != nil {
		return fmt.Errorf("failed to delete application-channel relationships: %w", err)
	}

	// Delete the channel
	if err := s.db.Delete(channel).Error; err != nil {
		return fmt.Errorf("failed to delete channel: %w", err)
	}

	return nil
}

// EnableChannelForApp enables a global channel for a specific application
func (s *ChannelService) EnableChannelForApp(appID, channelName string, req *models.ApplicationChannelRequest) (*models.ApplicationChannelResponse, error) {
	// Verify the channel exists
	_, err := s.GetChannelByName(channelName)
	if err != nil {
		return nil, fmt.Errorf("channel not found: %w", err)
	}

	// Check if relationship already exists
	var appChannel models.ApplicationChannel
	result := s.db.Where("app_id = ? AND channel_name = ?", appID, channelName).First(&appChannel)

	if result.Error == gorm.ErrRecordNotFound {
		// Create new relationship
		appChannel = models.ApplicationChannel{
			AppID:             appID,
			ChannelName:       channelName,
			IsEnabled:         req.IsEnabled,
			AutoPublish:       req.AutoPublish,
			RolloutPercentage: req.RolloutPercentage,
		}

		if err := s.db.Create(&appChannel).Error; err != nil {
			return nil, fmt.Errorf("failed to create application-channel relationship: %w", err)
		}
	} else if result.Error != nil {
		return nil, fmt.Errorf("failed to check application-channel relationship: %w", result.Error)
	} else {
		// Update existing relationship
		appChannel.IsEnabled = req.IsEnabled
		appChannel.AutoPublish = req.AutoPublish
		appChannel.RolloutPercentage = req.RolloutPercentage

		if err := s.db.Save(&appChannel).Error; err != nil {
			return nil, fmt.Errorf("failed to update application-channel relationship: %w", err)
		}
	}

	// Load the channel details for response
	if err := s.db.Preload("Channel").First(&appChannel, appChannel.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load channel details: %w", err)
	}

	return appChannel.ToResponse(), nil
}

// DisableChannelForApp disables a channel for a specific application
func (s *ChannelService) DisableChannelForApp(appID, channelName string) error {
	// Check if there are any published versions in this channel
	var versionCount int64
	if err := s.db.Model(&models.Version{}).
		Where("app_id = ? AND channel = ? AND is_published = ?", appID, channelName, true).
		Count(&versionCount).Error; err != nil {
		return fmt.Errorf("failed to check version count: %w", err)
	}

	if versionCount > 0 {
		return fmt.Errorf("cannot disable channel with published versions")
	}

	// Update the relationship to disabled
	if err := s.db.Model(&models.ApplicationChannel{}).
		Where("app_id = ? AND channel_name = ?", appID, channelName).
		Update("is_enabled", false).Error; err != nil {
		return fmt.Errorf("failed to disable channel: %w", err)
	}

	return nil
}

// InitializeDefaultChannels creates the standard global channels if they don't exist
func (s *ChannelService) InitializeDefaultChannels() error {
	defaultChannels := []models.Channel{
		{
			Name:        "stable",
			DisplayName: "Stable",
			Description: "Production-ready stable releases",
			IsActive:    true,
			SortOrder:   1,
		},
		{
			Name:        "beta",
			DisplayName: "Beta",
			Description: "Beta testing releases with new features",
			IsActive:    true,
			SortOrder:   2,
		},
		{
			Name:        "alpha",
			DisplayName: "Alpha",
			Description: "Alpha testing releases for early adopters",
			IsActive:    true,
			SortOrder:   3,
		},
	}

	for _, channel := range defaultChannels {
		var existingChannel models.Channel
		if err := s.db.Where("name = ?", channel.Name).First(&existingChannel).Error; err == gorm.ErrRecordNotFound {
			// Channel doesn't exist, create it
			if err := s.db.Create(&channel).Error; err != nil {
				return fmt.Errorf("failed to create default channel %s: %w", channel.Name, err)
			}
		}
	}

	return nil
}

// EnableDefaultChannelsForApp enables all default channels for a new application
func (s *ChannelService) EnableDefaultChannelsForApp(appID string) error {
	// Get all active global channels
	channels, err := s.GetAllChannels()
	if err != nil {
		return fmt.Errorf("failed to get channels: %w", err)
	}

	// Create application-channel relationships for each channel
	for _, channel := range channels {
		if !channel.IsActive {
			continue
		}

		// Check if relationship already exists
		var existingAppChannel models.ApplicationChannel
		if err := s.db.Where("app_id = ? AND channel_name = ?", appID, channel.Name).First(&existingAppChannel).Error; err == gorm.ErrRecordNotFound {
			// Create new relationship with default settings
			appChannel := models.ApplicationChannel{
				AppID:             appID,
				ChannelName:       channel.Name,
				IsEnabled:         true,
				AutoPublish:       false,
				RolloutPercentage: getRolloutPercentageByChannel(channel.Name),
			}

			if err := s.db.Create(&appChannel).Error; err != nil {
				return fmt.Errorf("failed to create application-channel relationship for %s: %w", channel.Name, err)
			}
		}
	}

	return nil
}

// Helper function to get default rollout percentage by channel
func getRolloutPercentageByChannel(channelName string) int {
	switch channelName {
	case "stable":
		return 100
	case "beta":
		return 50
	case "alpha":
		return 25
	default:
		return 100
	}
}

// ValidateChannelForApp checks if a channel is enabled for an application
func (s *ChannelService) ValidateChannelForApp(appID, channelName string) error {
	var appChannel models.ApplicationChannel
	if err := s.db.Where("app_id = ? AND channel_name = ? AND is_enabled = ?", appID, channelName, true).First(&appChannel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("channel %s is not enabled for this application", channelName)
		}
		return fmt.Errorf("failed to validate channel: %w", err)
	}
	return nil
}
