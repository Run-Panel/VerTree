package models

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Application represents an application in the system
type Application struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	AppID       string         `gorm:"uniqueIndex;column:app_id;size:32;not null" json:"app_id"`
	Name        string         `gorm:"uniqueIndex;size:100;not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Icon        string         `gorm:"column:icon;size:500" json:"icon_url"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedBy   uint           `gorm:"not null" json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Associations
	CreatedByAdmin Admin            `gorm:"foreignKey:CreatedBy" json:"-"`
	Versions       []Version        `gorm:"foreignKey:AppID;references:AppID" json:"-"`
	Channels       []Channel        `gorm:"many2many:application_channels;joinForeignKey:AppID;joinReferences:ChannelName;foreignKey:AppID;references:Name" json:"-"`
	Keys           []ApplicationKey `gorm:"foreignKey:AppID;references:AppID" json:"-"`
}

// ApplicationRequest represents the request for creating/updating an application
type ApplicationRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description"`
	Icon        string `json:"icon_url"`
	IsActive    bool   `json:"is_active"`
}

// ApplicationResponse represents the response format for an application
type ApplicationResponse struct {
	ID          uint      `json:"id"`
	AppID       string    `json:"app_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon_url"`
	IsActive    bool      `json:"is_active"`
	CreatedBy   uint      `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	KeysCount   int       `json:"keys_count"`
}

// ApplicationKey represents an API key for an application
type ApplicationKey struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	KeyID       string          `gorm:"uniqueIndex;column:key_id;size:32;not null" json:"key_id"`
	AppID       string          `gorm:"index;column:app_id;size:32;not null" json:"app_id"`
	Name        string          `gorm:"size:100;not null" json:"name"`
	KeySecret   string          `gorm:"column:key_secret;size:64;not null" json:"-"`
	KeyHash     string          `gorm:"index;column:key_hash;size:64;not null" json:"-"`
	Permissions PermissionsList `gorm:"type:json" json:"permissions"`
	IsActive    bool            `gorm:"default:true" json:"is_active"`
	LastUsed    *time.Time      `gorm:"column:last_used" json:"last_used"`
	CreatedBy   uint            `gorm:"not null" json:"created_by"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"index" json:"-"`

	// Associations
	Application    Application `gorm:"foreignKey:AppID;references:AppID" json:"-"`
	CreatedByAdmin Admin       `gorm:"foreignKey:CreatedBy" json:"-"`
}

// ApplicationKeyRequest represents the request for creating/updating an application key
type ApplicationKeyRequest struct {
	Name        string   `json:"name" validate:"required,min=2,max=100"`
	Permissions []string `json:"permissions" validate:"required,min=1"`
	IsActive    bool     `json:"is_active"`
}

// ApplicationKeyResponse represents the response format for an application key
type ApplicationKeyResponse struct {
	ID          uint            `json:"id"`
	KeyID       string          `json:"key_id"`
	AppID       string          `json:"app_id"`
	Name        string          `json:"name"`
	Permissions PermissionsList `json:"permissions"`
	IsActive    bool            `json:"is_active"`
	LastUsed    *time.Time      `json:"last_used"`
	CreatedBy   uint            `json:"created_by"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// ApplicationKeyCreatedResponse represents the response when a key is created (includes secret)
type ApplicationKeyCreatedResponse struct {
	ApplicationKeyResponse
	KeySecret string `json:"key_secret"`
}

// PermissionsList represents a list of permissions
type PermissionsList []string

// Value implements driver.Valuer interface for database storage
func (p PermissionsList) Value() (driver.Value, error) {
	if len(p) == 0 {
		return "[]", nil
	}
	data, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

// Scan implements sql.Scanner interface for database retrieval
func (p *PermissionsList) Scan(value interface{}) error {
	if value == nil {
		*p = PermissionsList{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		return fmt.Errorf("cannot scan %T into PermissionsList", value)
	}

	return json.Unmarshal(bytes, p)
}

// BeforeCreate sets the app_id before creating an application
func (a *Application) BeforeCreate(tx *gorm.DB) error {
	if a.AppID == "" {
		appID, err := generateAppID()
		if err != nil {
			return fmt.Errorf("failed to generate app ID: %w", err)
		}
		a.AppID = appID
	}
	return nil
}

// BeforeCreate sets the key_id before creating an application key
func (ak *ApplicationKey) BeforeCreate(tx *gorm.DB) error {
	if ak.KeyID == "" {
		keyID, err := generateKeyID()
		if err != nil {
			return fmt.Errorf("failed to generate key ID: %w", err)
		}
		ak.KeyID = keyID
	}
	return nil
}

// ToResponse converts Application to ApplicationResponse
func (a *Application) ToResponse() *ApplicationResponse {
	return &ApplicationResponse{
		ID:          a.ID,
		AppID:       a.AppID,
		Name:        a.Name,
		Description: a.Description,
		Icon:        a.Icon,
		IsActive:    a.IsActive,
		CreatedBy:   a.CreatedBy,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

// ToResponse converts ApplicationKey to ApplicationKeyResponse
func (ak *ApplicationKey) ToResponse() *ApplicationKeyResponse {
	return &ApplicationKeyResponse{
		ID:          ak.ID,
		KeyID:       ak.KeyID,
		AppID:       ak.AppID,
		Name:        ak.Name,
		Permissions: ak.Permissions,
		IsActive:    ak.IsActive,
		LastUsed:    ak.LastUsed,
		CreatedBy:   ak.CreatedBy,
		CreatedAt:   ak.CreatedAt,
		UpdatedAt:   ak.UpdatedAt,
	}
}

// generateAppID generates a unique application ID
func generateAppID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "app_" + hex.EncodeToString(bytes)[:28], nil
}

// generateKeyID generates a unique key ID
func generateKeyID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "key_" + hex.EncodeToString(bytes)[:28], nil
}

// GenerateKeySecret generates a random API key secret
func GenerateKeySecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
