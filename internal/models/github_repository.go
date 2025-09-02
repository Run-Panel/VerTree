package models

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"gorm.io/gorm"
)

// GitHubRepository represents the binding between an application and a GitHub repository
type GitHubRepository struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	AppID         string `json:"app_id" gorm:"size:32;not null;index" validate:"required"`
	RepositoryURL string `json:"repository_url" gorm:"not null;size:500" validate:"required,url"`
	OwnerName     string `json:"owner_name" gorm:"not null;size:100" validate:"required"`
	RepoName      string `json:"repo_name" gorm:"not null;size:100" validate:"required"`
	BranchName    string `json:"branch_name" gorm:"not null;size:100;default:main" validate:"required"`

	// Authentication fields
	AuthType    string `json:"auth_type" gorm:"size:20;default:token"` // "token" or "github_app"
	AccessToken string `json:"-" gorm:"size:500"`                      // Personal Access Token (for token auth)

	// GitHub Apps authentication fields
	GitHubAppID    int64  `json:"github_app_id" gorm:"default:0"`   // GitHub App ID
	InstallationID int64  `json:"installation_id" gorm:"default:0"` // Installation ID
	PrivateKey     string `json:"-" gorm:"type:text"`               // GitHub App private key (encrypted)

	WebhookSecret  string         `json:"-" gorm:"size:100"` // Secret for webhook validation
	WebhookID      int64          `json:"webhook_id" gorm:"default:0"`
	IsActive       bool           `json:"is_active" gorm:"default:true"`
	AutoSync       bool           `json:"auto_sync" gorm:"default:true"`
	AutoPublish    bool           `json:"auto_publish" gorm:"default:false"`
	DefaultChannel string         `json:"default_channel" gorm:"size:20;default:stable"`
	LastSyncAt     *time.Time     `json:"last_sync_at"`
	LastSyncStatus string         `json:"last_sync_status" gorm:"size:50;default:pending"`
	LastSyncError  string         `json:"last_sync_error" gorm:"type:text"`
	SyncCount      int64          `json:"sync_count" gorm:"default:0"`
	CreatedBy      uint           `json:"created_by" gorm:"not null"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// Associations
	Application    Application     `json:"application,omitempty" gorm:"foreignKey:AppID;references:AppID"`
	CreatedByAdmin Admin           `json:"created_by_admin,omitempty" gorm:"foreignKey:CreatedBy"`
	Releases       []GitHubRelease `json:"releases,omitempty" gorm:"foreignKey:RepositoryID"`
}

// TableName returns the table name for GitHubRepository model
func (GitHubRepository) TableName() string {
	return "github_repositories"
}

// GitHubRelease represents a synchronized GitHub release
type GitHubRelease struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	RepositoryID  uint           `json:"repository_id" gorm:"not null;index"`
	ReleaseID     int64          `json:"release_id" gorm:"not null;uniqueIndex:idx_repo_release"`
	TagName       string         `json:"tag_name" gorm:"not null;size:100"`
	ReleaseName   string         `json:"release_name" gorm:"not null;size:200"`
	Body          string         `json:"body" gorm:"type:text"`
	IsPrerelease  bool           `json:"is_prerelease" gorm:"default:false"`
	IsDraft       bool           `json:"is_draft" gorm:"default:false"`
	PublishedAt   *time.Time     `json:"published_at"`
	DownloadURL   string         `json:"download_url" gorm:"size:500"`
	FileSize      int64          `json:"file_size" gorm:"default:0"`
	FileChecksum  string         `json:"file_checksum" gorm:"size:128"`
	LocalFilePath string         `json:"local_file_path" gorm:"size:500"`
	SyncStatus    string         `json:"sync_status" gorm:"size:50;default:pending"`
	VersionID     *uint          `json:"version_id" gorm:"index"` // Linked to versions table
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// Associations
	Repository GitHubRepository `json:"repository,omitempty" gorm:"foreignKey:RepositoryID"`
	Version    *Version         `json:"version,omitempty" gorm:"foreignKey:VersionID"`
}

// TableName returns the table name for GitHubRelease model
func (GitHubRelease) TableName() string {
	return "github_releases"
}

// GitHubRepositoryRequest represents the request for creating/updating a GitHub repository binding
type GitHubRepositoryRequest struct {
	AppID         string `json:"app_id" validate:"required"`
	RepositoryURL string `json:"repository_url" validate:"required,url"`
	BranchName    string `json:"branch_name" validate:"required"`

	// Authentication configuration
	AuthType    string `json:"auth_type" validate:"required,oneof=token github_app"` // "token" or "github_app"
	AccessToken string `json:"access_token,omitempty"`                               // For token authentication

	// GitHub Apps configuration (required when auth_type=github_app)
	GitHubAppID    int64  `json:"github_app_id,omitempty"`
	InstallationID int64  `json:"installation_id,omitempty"`
	PrivateKey     string `json:"private_key,omitempty"` // GitHub App private key

	IsActive       bool   `json:"is_active"`
	AutoSync       bool   `json:"auto_sync"`
	AutoPublish    bool   `json:"auto_publish"`
	DefaultChannel string `json:"default_channel" validate:"required"`
}

// GitHubRepositoryResponse represents the response format for a GitHub repository binding
type GitHubRepositoryResponse struct {
	ID            uint   `json:"id"`
	AppID         string `json:"app_id"`
	RepositoryURL string `json:"repository_url"`
	OwnerName     string `json:"owner_name"`
	RepoName      string `json:"repo_name"`
	BranchName    string `json:"branch_name"`

	// Authentication information
	AuthType       string `json:"auth_type"`
	GitHubAppID    int64  `json:"github_app_id,omitempty"`
	InstallationID int64  `json:"installation_id,omitempty"`

	WebhookID      int64      `json:"webhook_id"`
	IsActive       bool       `json:"is_active"`
	AutoSync       bool       `json:"auto_sync"`
	AutoPublish    bool       `json:"auto_publish"`
	DefaultChannel string     `json:"default_channel"`
	LastSyncAt     *time.Time `json:"last_sync_at"`
	LastSyncStatus string     `json:"last_sync_status"`
	LastSyncError  string     `json:"last_sync_error"`
	SyncCount      int64      `json:"sync_count"`
	CreatedBy      uint       `json:"created_by"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	HasAccessToken bool       `json:"has_access_token"`
	HasGitHubApp   bool       `json:"has_github_app"`
}

// ToResponse converts GitHubRepository model to GitHubRepositoryResponse
func (gr *GitHubRepository) ToResponse() *GitHubRepositoryResponse {
	return &GitHubRepositoryResponse{
		ID:             gr.ID,
		AppID:          gr.AppID,
		RepositoryURL:  gr.RepositoryURL,
		OwnerName:      gr.OwnerName,
		RepoName:       gr.RepoName,
		BranchName:     gr.BranchName,
		AuthType:       gr.AuthType,
		GitHubAppID:    gr.GitHubAppID,
		InstallationID: gr.InstallationID,
		WebhookID:      gr.WebhookID,
		IsActive:       gr.IsActive,
		AutoSync:       gr.AutoSync,
		AutoPublish:    gr.AutoPublish,
		DefaultChannel: gr.DefaultChannel,
		LastSyncAt:     gr.LastSyncAt,
		LastSyncStatus: gr.LastSyncStatus,
		LastSyncError:  gr.LastSyncError,
		SyncCount:      gr.SyncCount,
		CreatedBy:      gr.CreatedBy,
		CreatedAt:      gr.CreatedAt,
		UpdatedAt:      gr.UpdatedAt,
		HasAccessToken: gr.AccessToken != "",
		HasGitHubApp:   gr.GitHubAppID > 0 && gr.PrivateKey != "",
	}
}

// GenerateWebhookSecret generates a random webhook secret
func (gr *GitHubRepository) GenerateWebhookSecret() error {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return err
	}
	gr.WebhookSecret = hex.EncodeToString(bytes)
	return nil
}

// SyncRequest represents a manual sync request
type SyncRequest struct {
	Force bool `json:"force"`
}

// SyncResponse represents a sync operation response
type SyncResponse struct {
	Status         string    `json:"status"`
	Message        string    `json:"message"`
	ReleasesFound  int       `json:"releases_found"`
	ReleasesSync   int       `json:"releases_sync"`
	VersionsCreate int       `json:"versions_create"`
	SyncedAt       time.Time `json:"synced_at"`
	Errors         []string  `json:"errors,omitempty"`
}
