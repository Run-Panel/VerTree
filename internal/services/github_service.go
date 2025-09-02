package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/Run-Panel/VerTree/internal/database"
	"github.com/Run-Panel/VerTree/internal/models"
	"gorm.io/gorm"
)

// GitHubService handles GitHub repository operations
type GitHubService struct {
	db             *gorm.DB
	versionService *VersionService
	fileService    *FileService
	githubAPI      *GitHubAPI
}

// NewGitHubService creates a new GitHub service instance
func NewGitHubService() *GitHubService {
	return &GitHubService{
		db:             database.DB,
		versionService: NewVersionService(),
		fileService:    NewFileService(),
		githubAPI:      NewGitHubAPI(),
	}
}

// CreateRepository creates a new GitHub repository binding
func (s *GitHubService) CreateRepository(req *models.GitHubRepositoryRequest, adminID uint) (*models.GitHubRepository, error) {
	// Parse repository URL
	ownerName, repoName, err := s.parseRepositoryURL(req.RepositoryURL)
	if err != nil {
		return nil, fmt.Errorf("invalid repository URL: %w", err)
	}

	// Check if repository binding already exists for this app
	var existingRepo models.GitHubRepository
	if err := s.db.Where("app_id = ? AND owner_name = ? AND repo_name = ?",
		req.AppID, ownerName, repoName).First(&existingRepo).Error; err == nil {
		return nil, fmt.Errorf("repository %s/%s is already bound to this application", ownerName, repoName)
	}

	// Create repository binding
	repo := &models.GitHubRepository{
		AppID:          req.AppID,
		RepositoryURL:  req.RepositoryURL,
		OwnerName:      ownerName,
		RepoName:       repoName,
		BranchName:     req.BranchName,
		AuthType:       req.AuthType,
		AccessToken:    req.AccessToken, // TODO: Encrypt this
		GitHubAppID:    req.GitHubAppID,
		InstallationID: req.InstallationID,
		PrivateKey:     req.PrivateKey, // TODO: Encrypt this
		IsActive:       req.IsActive,
		AutoSync:       req.AutoSync,
		AutoPublish:    req.AutoPublish,
		DefaultChannel: req.DefaultChannel,
		LastSyncStatus: "pending",
		CreatedBy:      adminID,
	}

	// Generate webhook secret
	if err := repo.GenerateWebhookSecret(); err != nil {
		return nil, fmt.Errorf("failed to generate webhook secret: %w", err)
	}

	if err := s.db.Create(repo).Error; err != nil {
		return nil, fmt.Errorf("failed to create repository binding: %w", err)
	}

	// Test GitHub connection and setup webhook
	token, err := s.getAccessToken(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	if token != "" {
		// Test connection first
		if err := s.githubAPI.TestConnection(ownerName, repoName, token); err != nil {
			return nil, fmt.Errorf("GitHub connection test failed: %w", err)
		}

		// Setup webhook
		if err := s.setupWebhook(repo); err != nil {
			// Log warning but don't fail the creation
			fmt.Printf("Warning: failed to setup webhook for %s/%s: %v\n", ownerName, repoName, err)
		}
	}

	return repo, nil
}

// GetRepositoryByID retrieves a repository binding by ID
func (s *GitHubService) GetRepositoryByID(id uint) (*models.GitHubRepository, error) {
	var repo models.GitHubRepository
	if err := s.db.Preload("Application").First(&repo, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("repository binding not found")
		}
		return nil, fmt.Errorf("failed to get repository binding: %w", err)
	}
	return &repo, nil
}

// GetRepositoriesByAppID retrieves all repository bindings for an app
func (s *GitHubService) GetRepositoriesByAppID(appID string) ([]*models.GitHubRepository, error) {
	var repos []*models.GitHubRepository
	if err := s.db.Where("app_id = ?", appID).Find(&repos).Error; err != nil {
		return nil, fmt.Errorf("failed to get repository bindings: %w", err)
	}
	return repos, nil
}

// GetAllRepositories retrieves all repository bindings
func (s *GitHubService) GetAllRepositories() ([]*models.GitHubRepository, error) {
	var repos []*models.GitHubRepository
	if err := s.db.Find(&repos).Error; err != nil {
		return nil, fmt.Errorf("failed to get all repository bindings: %w", err)
	}
	return repos, nil
}

// RepositoryValidationResult holds validation result with detailed information
type RepositoryValidationResult struct {
	Valid         bool                 `json:"valid"`
	Repository    *GitHubAPIRepository `json:"repository,omitempty"`
	LatestRelease *GitHubAPIRelease    `json:"latest_release,omitempty"`
	OwnerName     string               `json:"owner_name"`
	RepoName      string               `json:"repo_name"`
}

// ValidateRepositoryWithInfo validates a GitHub repository URL and returns detailed information
func (s *GitHubService) ValidateRepositoryWithInfo(repositoryURL, accessToken string) (*RepositoryValidationResult, error) {
	// Parse repository URL
	ownerName, repoName, err := s.parseRepositoryURL(repositoryURL)
	if err != nil {
		return nil, fmt.Errorf("invalid repository URL: %w", err)
	}

	// Get repository information from GitHub API
	repoInfo, err := s.githubAPI.GetRepository(ownerName, repoName, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository information: %w", err)
	}

	// Get latest release (optional)
	latestRelease, _ := s.githubAPI.GetLatestRelease(ownerName, repoName, accessToken)

	return &RepositoryValidationResult{
		Valid:         true,
		Repository:    repoInfo,
		LatestRelease: latestRelease,
		OwnerName:     ownerName,
		RepoName:      repoName,
	}, nil
}

// ValidateRepositoryWithGitHubApp validates a GitHub repository using GitHub Apps authentication
func (s *GitHubService) ValidateRepositoryWithGitHubApp(repositoryURL string, appID int64, privateKey string, installationID int64) (*RepositoryValidationResult, error) {
	// Parse repository URL
	ownerName, repoName, err := s.parseRepositoryURL(repositoryURL)
	if err != nil {
		return nil, fmt.Errorf("invalid repository URL: %w", err)
	}

	// Create GitHub Apps client
	appsAuth, err := NewGitHubAppsAuth(appID, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub Apps client: %w", err)
	}

	// Get installation token
	var token string
	if installationID > 0 {
		installationToken, err := appsAuth.GetInstallationToken(installationID)
		if err != nil {
			return nil, fmt.Errorf("failed to get installation token: %w", err)
		}
		token = installationToken.Token
	} else {
		token, err = appsAuth.GetTokenForRepository(ownerName, repoName)
		if err != nil {
			return nil, fmt.Errorf("failed to get token for repository: %w", err)
		}
	}

	// Get repository information from GitHub API
	repoInfo, err := s.githubAPI.GetRepository(ownerName, repoName, token)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository information: %w", err)
	}

	// Get latest release (optional)
	latestRelease, _ := s.githubAPI.GetLatestRelease(ownerName, repoName, token)

	return &RepositoryValidationResult{
		Valid:         true,
		Repository:    repoInfo,
		LatestRelease: latestRelease,
		OwnerName:     ownerName,
		RepoName:      repoName,
	}, nil
}

// UpdateRepository updates a repository binding
func (s *GitHubService) UpdateRepository(id uint, req *models.GitHubRepositoryRequest) (*models.GitHubRepository, error) {
	repo, err := s.GetRepositoryByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields
	repo.BranchName = req.BranchName
	repo.IsActive = req.IsActive
	repo.AutoSync = req.AutoSync
	repo.AutoPublish = req.AutoPublish
	repo.DefaultChannel = req.DefaultChannel
	repo.AuthType = req.AuthType

	// Update authentication credentials
	var needWebhookUpdate bool
	if req.AuthType == "token" && req.AccessToken != "" && req.AccessToken != repo.AccessToken {
		repo.AccessToken = req.AccessToken // TODO: Encrypt this
		repo.GitHubAppID = 0
		repo.InstallationID = 0
		repo.PrivateKey = ""
		needWebhookUpdate = true
	} else if req.AuthType == "github_app" {
		if req.GitHubAppID != repo.GitHubAppID || req.PrivateKey != repo.PrivateKey || req.InstallationID != repo.InstallationID {
			repo.GitHubAppID = req.GitHubAppID
			repo.InstallationID = req.InstallationID
			repo.PrivateKey = req.PrivateKey // TODO: Encrypt this
			repo.AccessToken = ""
			needWebhookUpdate = true
		}
	}

	// Re-setup webhook if credentials changed
	if needWebhookUpdate {
		if err := s.setupWebhook(repo); err != nil {
			fmt.Printf("Warning: failed to update webhook for %s/%s: %v\n", repo.OwnerName, repo.RepoName, err)
		}
	}

	if err := s.db.Save(repo).Error; err != nil {
		return nil, fmt.Errorf("failed to update repository binding: %w", err)
	}

	return repo, nil
}

// DeleteRepository deletes a repository binding
func (s *GitHubService) DeleteRepository(id uint) error {
	repo, err := s.GetRepositoryByID(id)
	if err != nil {
		return err
	}

	// Remove webhook if it exists
	if repo.WebhookID > 0 && repo.AccessToken != "" {
		if err := s.removeWebhook(repo); err != nil {
			fmt.Printf("Warning: failed to remove webhook for %s/%s: %v\n", repo.OwnerName, repo.RepoName, err)
		}
	}

	if err := s.db.Delete(repo).Error; err != nil {
		return fmt.Errorf("failed to delete repository binding: %w", err)
	}

	return nil
}

// SyncRepository manually synchronizes releases from GitHub
func (s *GitHubService) SyncRepository(id uint, force bool) (*models.SyncResponse, error) {
	repo, err := s.GetRepositoryByID(id)
	if err != nil {
		return nil, err
	}

	if !repo.IsActive {
		return nil, fmt.Errorf("repository binding is not active")
	}

	// Get access token based on authentication type
	token, err := s.getAccessToken(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}
	if token == "" {
		return nil, fmt.Errorf("access token is required for synchronization")
	}

	// Update sync status
	repo.LastSyncStatus = "syncing"
	repo.LastSyncError = ""
	s.db.Save(repo)

	response := &models.SyncResponse{
		Status:   "success",
		SyncedAt: time.Now(),
		Errors:   []string{},
	}

	// Fetch releases from GitHub
	releases, err := s.fetchGitHubReleases(repo)
	if err != nil {
		repo.LastSyncStatus = "failed"
		repo.LastSyncError = err.Error()
		s.db.Save(repo)
		return nil, fmt.Errorf("failed to fetch GitHub releases: %w", err)
	}

	response.ReleasesFound = len(releases)

	// Process each release
	for _, release := range releases {
		if err := s.processRelease(repo, release, force); err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("Release %s: %v", release.Release.TagName, err))
			continue
		}
		response.ReleasesSync++
	}

	// Update repository sync status
	repo.LastSyncAt = &response.SyncedAt
	repo.SyncCount++
	if len(response.Errors) == 0 {
		repo.LastSyncStatus = "success"
		repo.LastSyncError = ""
	} else {
		repo.LastSyncStatus = "partial"
		repo.LastSyncError = fmt.Sprintf("Processed %d/%d releases successfully",
			response.ReleasesSync, response.ReleasesFound)
	}
	s.db.Save(repo)

	response.Message = fmt.Sprintf("Synchronized %d/%d releases", response.ReleasesSync, response.ReleasesFound)
	return response, nil
}

// ProcessWebhook processes GitHub webhook payload
func (s *GitHubService) ProcessWebhook(payload []byte, signature string, repoID uint) error {
	repo, err := s.GetRepositoryByID(repoID)
	if err != nil {
		return err
	}

	// Verify webhook signature
	if !s.verifyWebhookSignature(payload, signature, repo.WebhookSecret) {
		return fmt.Errorf("invalid webhook signature")
	}

	// Parse webhook payload
	var webhookPayload WebhookPayload
	if err := json.Unmarshal(payload, &webhookPayload); err != nil {
		return fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	// Only process release events
	if webhookPayload.Action == "published" || webhookPayload.Action == "released" {
		return s.processWebhookRelease(repo, &webhookPayload)
	}

	return nil
}

// Private helper methods

func (s *GitHubService) parseRepositoryURL(repositoryURL string) (string, string, error) {
	// Parse different GitHub URL formats
	// https://github.com/owner/repo
	// https://github.com/owner/repo.git
	// git@github.com:owner/repo.git

	var ownerName, repoName string

	if strings.HasPrefix(repositoryURL, "git@github.com:") {
		// SSH format: git@github.com:owner/repo.git
		path := strings.TrimPrefix(repositoryURL, "git@github.com:")
		path = strings.TrimSuffix(path, ".git")
		parts := strings.Split(path, "/")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid SSH repository URL format")
		}
		ownerName, repoName = parts[0], parts[1]
	} else {
		// HTTPS format: https://github.com/owner/repo or https://github.com/owner/repo.git
		u, err := url.Parse(repositoryURL)
		if err != nil {
			return "", "", fmt.Errorf("invalid repository URL: %w", err)
		}

		if u.Host != "github.com" {
			return "", "", fmt.Errorf("only GitHub repositories are supported")
		}

		path := strings.Trim(u.Path, "/")
		path = strings.TrimSuffix(path, ".git")
		parts := strings.Split(path, "/")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid repository URL format")
		}
		ownerName, repoName = parts[0], parts[1]
	}

	// Validate owner and repo names
	if ownerName == "" || repoName == "" {
		return "", "", fmt.Errorf("owner and repository names cannot be empty")
	}

	// Basic validation for GitHub naming conventions
	nameRegex := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !nameRegex.MatchString(ownerName) || !nameRegex.MatchString(repoName) {
		return "", "", fmt.Errorf("invalid owner or repository name format")
	}

	return ownerName, repoName, nil
}

func (s *GitHubService) setupWebhook(repo *models.GitHubRepository) error {
	// Get access token based on authentication type
	token, err := s.getAccessToken(repo)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}
	if token == "" {
		return fmt.Errorf("access token is required to setup webhook")
	}

	// Build webhook URL (this should come from config)
	webhookURL := fmt.Sprintf("https://your-domain.com/api/v1/webhook/github/%d", repo.ID)

	// Create or update webhook
	if repo.WebhookID > 0 {
		// Update existing webhook
		webhook, err := s.githubAPI.UpdateWebhook(repo.OwnerName, repo.RepoName,
			token, repo.WebhookID, webhookURL, repo.WebhookSecret)
		if err != nil {
			return fmt.Errorf("failed to update webhook: %w", err)
		}
		repo.WebhookID = webhook.ID
	} else {
		// Create new webhook
		webhook, err := s.githubAPI.CreateWebhook(repo.OwnerName, repo.RepoName,
			token, webhookURL, repo.WebhookSecret)
		if err != nil {
			return fmt.Errorf("failed to create webhook: %w", err)
		}
		repo.WebhookID = webhook.ID
	}

	// Save updated webhook ID
	return s.db.Save(repo).Error
}

func (s *GitHubService) removeWebhook(repo *models.GitHubRepository) error {
	if repo.WebhookID == 0 {
		return nil // Nothing to remove
	}

	// Get access token based on authentication type
	token, err := s.getAccessToken(repo)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}
	if token == "" {
		return nil // Cannot remove webhook without token
	}

	if err := s.githubAPI.DeleteWebhook(repo.OwnerName, repo.RepoName,
		token, repo.WebhookID); err != nil {
		return fmt.Errorf("failed to remove webhook: %w", err)
	}

	// Clear webhook ID
	repo.WebhookID = 0
	return s.db.Save(repo).Error
}

func (s *GitHubService) fetchGitHubReleases(repo *models.GitHubRepository) ([]*WebhookPayload, error) {
	// Get access token based on authentication type
	token, err := s.getAccessToken(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}
	if token == "" {
		return nil, fmt.Errorf("access token is required to fetch releases")
	}

	// Fetch releases from GitHub API (first page, up to 30 releases)
	apiReleases, err := s.githubAPI.GetReleases(repo.OwnerName, repo.RepoName,
		token, 30, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases from GitHub: %w", err)
	}

	// Convert API releases to webhook payload format
	var releases []*WebhookPayload
	for _, apiRelease := range apiReleases {
		payload := ConvertToWebhookPayload(&apiRelease, repo.OwnerName, repo.RepoName)
		releases = append(releases, payload)
	}

	return releases, nil
}

func (s *GitHubService) processRelease(repo *models.GitHubRepository, release *WebhookPayload, force bool) error {
	// Check if release already exists
	var existingRelease models.GitHubRelease
	if err := s.db.Where("repository_id = ? AND release_id = ?",
		repo.ID, release.Release.ID).First(&existingRelease).Error; err == nil {
		if !force {
			return nil // Release already processed
		}
	}

	// Create or update GitHub release record
	ghRelease := &models.GitHubRelease{
		RepositoryID: repo.ID,
		ReleaseID:    release.Release.ID,
		TagName:      release.Release.TagName,
		ReleaseName:  release.Release.Name,
		Body:         release.Release.Body,
		IsPrerelease: release.Release.Prerelease,
		IsDraft:      release.Release.Draft,
		PublishedAt:  &release.Release.PublishedAt,
		SyncStatus:   "processing",
	}

	// Find the best download asset
	if len(release.Release.Assets) > 0 {
		asset := s.selectBestAsset(release.Release.Assets)
		ghRelease.DownloadURL = asset.BrowserDownloadURL
		ghRelease.FileSize = asset.Size
	}

	if existingRelease.ID > 0 {
		ghRelease.ID = existingRelease.ID
		s.db.Save(ghRelease)
	} else {
		s.db.Create(ghRelease)
	}

	// Create version if auto-sync is enabled
	if repo.AutoSync {
		if err := s.createVersionFromRelease(repo, ghRelease); err != nil {
			ghRelease.SyncStatus = "failed"
			s.db.Save(ghRelease)
			return fmt.Errorf("failed to create version: %w", err)
		}
	}

	ghRelease.SyncStatus = "completed"
	s.db.Save(ghRelease)

	return nil
}

func (s *GitHubService) selectBestAsset(assets []struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}) struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
} {
	// Simple asset selection logic - prefer archives and executables
	preferredExtensions := []string{".zip", ".tar.gz", ".tar.bz2", ".exe", ".dmg", ".deb", ".rpm"}

	for _, ext := range preferredExtensions {
		for _, asset := range assets {
			if strings.HasSuffix(strings.ToLower(asset.Name), ext) {
				return asset
			}
		}
	}

	// Return first asset if no preferred type found
	if len(assets) > 0 {
		return assets[0]
	}

	return struct {
		ID                 int64  `json:"id"`
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	}{}
}

func (s *GitHubService) createVersionFromRelease(repo *models.GitHubRepository, release *models.GitHubRelease) error {
	// Check if version already exists
	var existingVersion models.Version
	if err := s.db.Where("app_id = ? AND version = ?",
		repo.AppID, release.TagName).First(&existingVersion).Error; err == nil {
		// Link existing version to release
		release.VersionID = &existingVersion.ID
		s.db.Save(release)
		return nil
	}

	// Download and cache file if URL is available
	var fileURL, fileChecksum string
	var fileSize int64

	if release.DownloadURL != "" {
		cachedFile, err := s.fileService.CacheGitHubAsset(release.DownloadURL, repo.AppID, release.TagName)
		if err != nil {
			return fmt.Errorf("failed to cache release asset: %w", err)
		}
		fileURL = cachedFile.URL
		fileChecksum = cachedFile.Checksum
		fileSize = cachedFile.Size
		release.LocalFilePath = cachedFile.LocalPath
		release.FileChecksum = fileChecksum
	} else {
		// Use original download URL if caching fails
		fileURL = release.DownloadURL
		fileSize = release.FileSize
	}

	// Determine channel based on release type
	channel := repo.DefaultChannel
	if release.IsPrerelease {
		channel = "beta"
	}

	// Create version request
	versionReq := &models.VersionRequest{
		AppID:        repo.AppID,
		Version:      release.TagName,
		Channel:      channel,
		Title:        release.ReleaseName,
		Description:  release.Body,
		ReleaseNotes: release.Body,
		FileURL:      fileURL,
		FileSize:     fileSize,
		FileChecksum: fileChecksum,
		IsForced:     false,
	}

	// Create version
	version, err := s.versionService.CreateVersion(versionReq)
	if err != nil {
		return err
	}

	// Link release to version
	release.VersionID = &version.ID
	s.db.Save(release)

	// Auto-publish if enabled
	if repo.AutoPublish && !release.IsPrerelease {
		if _, err := s.versionService.PublishVersion(version.ID); err != nil {
			fmt.Printf("Warning: failed to auto-publish version %s: %v\n", version.Version, err)
		}
	}

	return nil
}

func (s *GitHubService) processWebhookRelease(repo *models.GitHubRepository, payload *WebhookPayload) error {
	// Process release from webhook
	return s.processRelease(repo, payload, false)
}

func (s *GitHubService) verifyWebhookSignature(payload []byte, signature, secret string) bool {
	// GitHub sends signature in format "sha256=<signature>"
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}

	expectedSignature := strings.TrimPrefix(signature, "sha256=")

	// Calculate HMAC
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	calculatedSignature := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(expectedSignature), []byte(calculatedSignature))
}

// getAccessToken returns the appropriate access token based on authentication type
func (s *GitHubService) getAccessToken(repo *models.GitHubRepository) (string, error) {
	switch repo.AuthType {
	case "token":
		return repo.AccessToken, nil

	case "github_app":
		if repo.GitHubAppID == 0 || repo.PrivateKey == "" {
			return "", fmt.Errorf("GitHub App credentials not configured")
		}

		// Create GitHub Apps client
		appsAuth, err := NewGitHubAppsAuth(repo.GitHubAppID, repo.PrivateKey)
		if err != nil {
			return "", fmt.Errorf("failed to create GitHub Apps client: %w", err)
		}

		// Get installation token
		if repo.InstallationID > 0 {
			// Use specific installation ID
			token, err := appsAuth.GetInstallationToken(repo.InstallationID)
			if err != nil {
				return "", fmt.Errorf("failed to get installation token: %w", err)
			}
			return token.Token, nil
		} else {
			// Find installation for repository
			token, err := appsAuth.GetTokenForRepository(repo.OwnerName, repo.RepoName)
			if err != nil {
				return "", fmt.Errorf("failed to get token for repository: %w", err)
			}
			return token, nil
		}

	default:
		return "", fmt.Errorf("unsupported authentication type: %s", repo.AuthType)
	}
}
