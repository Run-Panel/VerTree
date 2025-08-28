package services

import (
	"fmt"

	"github.com/Run-Panel/VerTree/internal/database"
	"github.com/Run-Panel/VerTree/internal/models"
	"github.com/Run-Panel/VerTree/internal/utils"
	"gorm.io/gorm"
)

// UpdateService handles update check business logic
type UpdateService struct {
	db         *gorm.DB
	versionSvc *VersionService
	channelSvc *ChannelService
	statsSvc   *StatsService
	versionCmp *utils.VersionComparer
}

// NewUpdateService creates a new update service instance
func NewUpdateService() *UpdateService {
	return &UpdateService{
		db:         database.DB,
		versionSvc: NewVersionService(),
		channelSvc: NewChannelService(),
		statsSvc:   NewStatsService(),
		versionCmp: utils.NewVersionComparer(),
	}
}

// CheckUpdate checks for available updates
func (s *UpdateService) CheckUpdate(req *models.CheckUpdateRequest, clientIP string) (*models.CheckUpdateResponse, error) {
	// Validate that the app exists and is active
	var app models.Application
	if err := s.db.Where("app_id = ? AND is_active = ?", req.AppID, true).First(&app).Error; err != nil {
		return nil, fmt.Errorf("application not found or inactive: %s", req.AppID)
	}

	// Record the check action
	statReq := &models.UpdateStatRequest{
		Version:       req.CurrentVersion,
		ClientID:      req.ClientID,
		ClientVersion: req.CurrentVersion,
		Region:        req.Region,
		Action:        "check",
	}

	// Record stat asynchronously (don't fail the request if this fails)
	go func() {
		if err := s.statsSvc.RecordUpdateStat(statReq, clientIP); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to record update stat: %v\n", err)
		}
	}()

	// Validate channel for this specific app
	channel, err := s.channelSvc.GetChannelByAppAndName(req.AppID, req.Channel)
	if err != nil {
		return nil, fmt.Errorf("invalid channel %s for app %s", req.Channel, req.AppID)
	}

	if !channel.IsActive {
		return nil, fmt.Errorf("channel %s is not active", req.Channel)
	}

	// Get latest version for the app and channel
	latestVersion, err := s.versionSvc.GetLatestVersionForApp(req.AppID, req.Channel)
	if err != nil {
		// No published version available
		return &models.CheckUpdateResponse{
			HasUpdate: false,
		}, nil
	}

	// Check if update is needed
	hasUpdate := s.isUpdateNeeded(req.CurrentVersion, latestVersion.Version)
	if !hasUpdate {
		return &models.CheckUpdateResponse{
			HasUpdate: false,
		}, nil
	}

	// Check if client meets minimum upgrade requirements
	if latestVersion.MinUpgradeVersion != "" {
		if !s.meetsMinimumVersion(req.CurrentVersion, latestVersion.MinUpgradeVersion) {
			return &models.CheckUpdateResponse{
				HasUpdate: false,
			}, nil
		}
	}

	// Check rollout rules
	if !s.shouldReceiveUpdate(req, channel) {
		return &models.CheckUpdateResponse{
			HasUpdate: false,
		}, nil
	}

	// Build response with update information
	response := &models.CheckUpdateResponse{
		HasUpdate:         true,
		LatestVersion:     latestVersion.Version,
		DownloadURL:       s.buildDownloadURL(latestVersion.FileURL, req),
		FileSize:          latestVersion.FileSize,
		FileChecksum:      latestVersion.FileChecksum,
		IsForced:          latestVersion.IsForced,
		Title:             latestVersion.Title,
		Description:       latestVersion.Description,
		ReleaseNotes:      latestVersion.ReleaseNotes,
		MinUpgradeVersion: latestVersion.MinUpgradeVersion,
	}

	return response, nil
}

// isUpdateNeeded checks if an update is needed based on semantic version comparison
func (s *UpdateService) isUpdateNeeded(currentVersion, latestVersion string) bool {
	return s.versionCmp.IsUpdateNeeded(currentVersion, latestVersion)
}

// meetsMinimumVersion checks if the current version meets the minimum upgrade version
func (s *UpdateService) meetsMinimumVersion(currentVersion, minVersion string) bool {
	return s.versionCmp.MeetsMinimumVersion(currentVersion, minVersion)
}

// shouldReceiveUpdate checks if the client should receive the update based on rollout rules
func (s *UpdateService) shouldReceiveUpdate(req *models.CheckUpdateRequest, channel *models.Channel) bool {
	// Check rollout percentage
	if channel.RolloutPercentage < 100 {
		// Simple hash-based rollout - use client ID to determine eligibility
		// This ensures consistent behavior for the same client
		hash := s.hashString(req.ClientID)
		if hash%100 >= channel.RolloutPercentage {
			return false
		}
	}

	// You can add more complex rollout rules here based on region, client version, etc.

	return true
}

// buildDownloadURL builds the download URL based on client requirements
func (s *UpdateService) buildDownloadURL(baseURL string, req *models.CheckUpdateRequest) string {
	// For now, return the base URL as-is
	// In production, you might want to:
	// 1. Choose different CDN endpoints based on region
	// 2. Add architecture-specific paths
	// 3. Add authentication tokens

	return baseURL
}

// hashString creates a simple hash from a string for rollout calculations
func (s *UpdateService) hashString(str string) int {
	hash := 0
	for _, c := range str {
		hash = 31*hash + int(c)
		if hash < 0 {
			hash = -hash
		}
	}
	return hash
}

// RecordDownloadStart records when a download starts
func (s *UpdateService) RecordDownloadStart(version, clientID string, clientIP string) error {
	statReq := &models.UpdateStatRequest{
		Version:  version,
		ClientID: clientID,
		Action:   "download",
	}

	return s.statsSvc.RecordUpdateStat(statReq, clientIP)
}

// RecordInstallResult records the result of an installation
func (s *UpdateService) RecordInstallResult(version, clientID string, success bool, errorMessage string, clientIP string) error {
	action := "success"
	if !success {
		action = "failed"
	}

	statReq := &models.UpdateStatRequest{
		Version:      version,
		ClientID:     clientID,
		Action:       action,
		ErrorMessage: errorMessage,
	}

	return s.statsSvc.RecordUpdateStat(statReq, clientIP)
}
