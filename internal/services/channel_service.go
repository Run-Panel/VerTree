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

// GetAllChannels retrieves all channels
func (s *ChannelService) GetAllChannels() ([]*models.Channel, error) {
	var channels []*models.Channel
	if err := s.db.Order("name").Find(&channels).Error; err != nil {
		return nil, fmt.Errorf("failed to get channels: %w", err)
	}
	return channels, nil
}

// GetChannelByID retrieves a channel by ID
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

// GetChannelByName retrieves a channel by name
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

// UpdateChannel updates an existing channel
func (s *ChannelService) UpdateChannel(id uint, req *models.ChannelRequest) (*models.Channel, error) {
	channel, err := s.GetChannelByID(id)
	if err != nil {
		return nil, err
	}

	// Check if new name conflicts (if changed)
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
	channel.AutoPublish = req.AutoPublish
	channel.RolloutPercentage = req.RolloutPercentage

	if err := s.db.Save(channel).Error; err != nil {
		return nil, fmt.Errorf("failed to update channel: %w", err)
	}

	return channel, nil
}

// CreateChannel creates a new channel
func (s *ChannelService) CreateChannel(req *models.ChannelRequest) (*models.Channel, error) {
	// Check if channel name already exists
	var existingChannel models.Channel
	if err := s.db.Where("name = ?", req.Name).First(&existingChannel).Error; err == nil {
		return nil, fmt.Errorf("channel name %s already exists", req.Name)
	}

	channel := &models.Channel{
		Name:              req.Name,
		DisplayName:       req.DisplayName,
		Description:       req.Description,
		IsActive:          req.IsActive,
		AutoPublish:       req.AutoPublish,
		RolloutPercentage: req.RolloutPercentage,
	}

	if err := s.db.Create(channel).Error; err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	return channel, nil
}

// DeleteChannel deletes a channel
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

	if err := s.db.Delete(channel).Error; err != nil {
		return fmt.Errorf("failed to delete channel: %w", err)
	}

	return nil
}
