package services

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/Run-Panel/VerTree/internal/database"
	"github.com/Run-Panel/VerTree/internal/models"
	"gorm.io/gorm"
)

// ApplicationService handles application-related business logic
type ApplicationService struct {
	db *gorm.DB
}

// NewApplicationService creates a new application service instance
func NewApplicationService() *ApplicationService {
	return &ApplicationService{
		db: database.DB,
	}
}

// CreateApplication creates a new application
func (s *ApplicationService) CreateApplication(req *models.ApplicationRequest, adminID uint) (*models.Application, error) {
	// Check if application with the same name already exists
	var existingApp models.Application
	if err := s.db.Where("name = ?", req.Name).First(&existingApp).Error; err == nil {
		return nil, fmt.Errorf("application with name '%s' already exists", req.Name)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing application: %w", err)
	}

	app := &models.Application{
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		IsActive:    req.IsActive,
		CreatedBy:   adminID,
	}

	if err := s.db.Create(app).Error; err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	// Enable default channels for the application
	channelService := NewChannelService()
	if err := channelService.EnableDefaultChannelsForApp(app.AppID); err != nil {
		// Log error but don't fail the application creation
		fmt.Printf("Warning: failed to enable default channels for app %s: %v\n", app.AppID, err)
	}

	return app, nil
}

// GetApplications retrieves applications with pagination
func (s *ApplicationService) GetApplications(page, limit int, adminID uint) (*models.PaginatedResponse, error) {
	var applications []models.Application
	var total int64

	// Count total applications
	query := s.db.Model(&models.Application{})
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count applications: %w", err)
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Fetch applications with pagination
	if err := s.db.Preload("CreatedByAdmin").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&applications).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch applications: %w", err)
	}

	// Convert to response format
	var responses []*models.ApplicationResponse
	for _, app := range applications {
		// Get keys count
		var keysCount int64
		s.db.Model(&models.ApplicationKey{}).Where("app_id = ?", app.AppID).Count(&keysCount)

		response := app.ToResponse()
		response.KeysCount = int(keysCount)
		responses = append(responses, response)
	}

	// Calculate pagination info
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	hasNext := page < totalPages
	hasPrev := page > 1

	pagination := models.PaginationResponse{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}

	return &models.PaginatedResponse{
		Code:       200,
		Message:    "success",
		Data:       responses,
		Pagination: pagination,
	}, nil
}

// GetApplication retrieves a single application by ID or AppID
func (s *ApplicationService) GetApplication(identifier string) (*models.Application, error) {
	var app models.Application

	// Try to find by AppID first, then by ID
	err := s.db.Preload("CreatedByAdmin").
		Where("app_id = ? OR id = ?", identifier, identifier).
		First(&app).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("application not found")
		}
		return nil, fmt.Errorf("failed to fetch application: %w", err)
	}

	return &app, nil
}

// UpdateApplication updates an existing application
func (s *ApplicationService) UpdateApplication(appID string, req *models.ApplicationRequest, adminID uint) (*models.Application, error) {
	var app models.Application

	// Find the application
	if err := s.db.Where("app_id = ?", appID).First(&app).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("application not found")
		}
		return nil, fmt.Errorf("failed to find application: %w", err)
	}

	// Check if another application with the same name exists
	if req.Name != app.Name {
		var existingApp models.Application
		if err := s.db.Where("name = ? AND app_id != ?", req.Name, appID).First(&existingApp).Error; err == nil {
			return nil, fmt.Errorf("application with name '%s' already exists", req.Name)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to check existing application: %w", err)
		}
	}

	// Update fields
	app.Name = req.Name
	app.Description = req.Description
	app.Icon = req.Icon
	app.IsActive = req.IsActive

	if err := s.db.Save(&app).Error; err != nil {
		return nil, fmt.Errorf("failed to update application: %w", err)
	}

	return &app, nil
}

// DeleteApplication soft deletes an application and its related data
func (s *ApplicationService) DeleteApplication(appID string, adminID uint) error {
	var app models.Application

	// Find the application
	if err := s.db.Where("app_id = ?", appID).First(&app).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("application not found")
		}
		return fmt.Errorf("failed to find application: %w", err)
	}

	// Begin transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Soft delete related data
	if err := tx.Where("app_id = ?", appID).Delete(&models.Version{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete versions: %w", err)
	}

	if err := tx.Where("app_id = ?", appID).Delete(&models.Channel{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete channels: %w", err)
	}

	if err := tx.Where("app_id = ?", appID).Delete(&models.ApplicationKey{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete application keys: %w", err)
	}

	// Soft delete the application
	if err := tx.Delete(&app).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete application: %w", err)
	}

	return tx.Commit().Error
}

// CreateApplicationKey creates a new API key for an application
func (s *ApplicationService) CreateApplicationKey(appID string, req *models.ApplicationKeyRequest, adminID uint) (*models.ApplicationKeyCreatedResponse, error) {
	// Verify application exists
	var app models.Application
	if err := s.db.Where("app_id = ?", appID).First(&app).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("application not found")
		}
		return nil, fmt.Errorf("failed to find application: %w", err)
	}

	// Check if key with the same name already exists for this app
	var existingKey models.ApplicationKey
	if err := s.db.Where("app_id = ? AND name = ?", appID, req.Name).First(&existingKey).Error; err == nil {
		return nil, fmt.Errorf("key with name '%s' already exists for this application", req.Name)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing key: %w", err)
	}

	// Generate key secret
	keySecret, err := models.GenerateKeySecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key secret: %w", err)
	}

	// Create SHA256 hash of the key for validation
	hasher := sha256.New()
	hasher.Write([]byte(keySecret))
	keyHash := hex.EncodeToString(hasher.Sum(nil))

	// Set default permissions if not provided
	permissions := req.Permissions
	if len(permissions) == 0 {
		permissions = []string{"check_update", "download", "install"}
	}

	key := &models.ApplicationKey{
		AppID:       appID,
		Name:        req.Name,
		KeySecret:   keySecret,
		KeyHash:     keyHash,
		Permissions: models.PermissionsList(permissions),
		IsActive:    req.IsActive,
		CreatedBy:   adminID,
	}

	if err := s.db.Create(key).Error; err != nil {
		return nil, fmt.Errorf("failed to create application key: %w", err)
	}

	// Return response with the secret (only shown once)
	response := &models.ApplicationKeyCreatedResponse{
		ApplicationKeyResponse: *key.ToResponse(),
		KeySecret:              keySecret,
	}

	return response, nil
}

// GetApplicationKeys retrieves all keys for an application
func (s *ApplicationService) GetApplicationKeys(appID string, adminID uint) ([]*models.ApplicationKeyResponse, error) {
	// Verify application exists
	var app models.Application
	if err := s.db.Where("app_id = ?", appID).First(&app).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("application not found")
		}
		return nil, fmt.Errorf("failed to find application: %w", err)
	}

	var keys []models.ApplicationKey
	if err := s.db.Where("app_id = ?", appID).
		Order("created_at DESC").
		Find(&keys).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch application keys: %w", err)
	}

	var responses []*models.ApplicationKeyResponse
	for _, key := range keys {
		responses = append(responses, key.ToResponse())
	}

	return responses, nil
}

// UpdateApplicationKey updates an existing application key
func (s *ApplicationService) UpdateApplicationKey(appID, keyID string, req *models.ApplicationKeyRequest, adminID uint) (*models.ApplicationKeyResponse, error) {
	var key models.ApplicationKey

	// Find the key
	if err := s.db.Where("app_id = ? AND key_id = ?", appID, keyID).First(&key).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("application key not found")
		}
		return nil, fmt.Errorf("failed to find application key: %w", err)
	}

	// Check if another key with the same name exists
	if req.Name != key.Name {
		var existingKey models.ApplicationKey
		if err := s.db.Where("app_id = ? AND name = ? AND key_id != ?", appID, req.Name, keyID).First(&existingKey).Error; err == nil {
			return nil, fmt.Errorf("key with name '%s' already exists for this application", req.Name)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to check existing key: %w", err)
		}
	}

	// Update fields
	key.Name = req.Name
	key.Permissions = models.PermissionsList(req.Permissions)
	key.IsActive = req.IsActive

	if err := s.db.Save(&key).Error; err != nil {
		return nil, fmt.Errorf("failed to update application key: %w", err)
	}

	return key.ToResponse(), nil
}

// DeleteApplicationKey soft deletes an application key
func (s *ApplicationService) DeleteApplicationKey(appID, keyID string, adminID uint) error {
	var key models.ApplicationKey

	// Find the key
	if err := s.db.Where("app_id = ? AND key_id = ?", appID, keyID).First(&key).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("application key not found")
		}
		return fmt.Errorf("failed to find application key: %w", err)
	}

	if err := s.db.Delete(&key).Error; err != nil {
		return fmt.Errorf("failed to delete application key: %w", err)
	}

	return nil
}

// ValidateAPIKey validates an API key and returns the associated application and key
func (s *ApplicationService) ValidateAPIKey(appID, keySecret string) (*models.Application, *models.ApplicationKey, error) {
	// Find the application
	var app models.Application
	if err := s.db.Where("app_id = ? AND is_active = ?", appID, true).First(&app).Error; err != nil {
		return nil, nil, fmt.Errorf("invalid application ID")
	}

	// Create hash of the provided key
	hasher := sha256.New()
	hasher.Write([]byte(keySecret))
	keyHash := hex.EncodeToString(hasher.Sum(nil))

	// Find the key by hash
	var key models.ApplicationKey
	if err := s.db.Where("app_id = ? AND key_hash = ? AND is_active = ?", appID, keyHash, true).First(&key).Error; err != nil {
		return nil, nil, fmt.Errorf("invalid API key")
	}

	// Update last used timestamp
	now := time.Now()
	key.LastUsed = &now
	s.db.Save(&key)

	return &app, &key, nil
}
