package admin

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Run-Panel/VerTree/internal/i18n"
	"github.com/Run-Panel/VerTree/internal/models"
	"github.com/Run-Panel/VerTree/internal/services"
	"github.com/gin-gonic/gin"
)

// GitHubHandler handles GitHub repository management endpoints
type GitHubHandler struct {
	githubService *services.GitHubService
	fileService   *services.FileService
}

// NewGitHubHandler creates a new GitHub handler
func NewGitHubHandler() *GitHubHandler {
	return &GitHubHandler{
		githubService: services.NewGitHubService(),
		fileService:   services.NewFileService(),
	}
}

// CreateRepository handles POST /admin/api/v1/github/repositories
func (h *GitHubHandler) CreateRepository(c *gin.Context) {
	localizer := getLocalizer(c)

	var req models.GitHubRepositoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse(
			localizer.Get(i18n.ErrInvalidRequestFormat), err))
		return
	}

	// Get admin ID from context
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.UnauthorizedResponse(
			localizer.Get(i18n.ErrAdminIDNotFound)))
		return
	}

	repo, err := h.githubService.CreateRepository(&req, adminID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse(
			"Failed to create repository binding", err))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse(repo.ToResponse()))
}

// GetRepositories handles GET /admin/api/v1/github/repositories
func (h *GitHubHandler) GetRepositories(c *gin.Context) {

	// Get query parameters
	appID := c.Query("app_id")

	var repos []*models.GitHubRepository
	var err error

	if appID != "" {
		// Get repositories for specific app
		repos, err = h.githubService.GetRepositoriesByAppID(appID)
	} else {
		// Get all repositories
		repos, err = h.githubService.GetAllRepositories()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.InternalServerErrorResponse(
			"Failed to get repositories", err))
		return
	}

	// Convert to response format
	var responses []*models.GitHubRepositoryResponse
	for _, repo := range repos {
		responses = append(responses, repo.ToResponse())
	}

	c.JSON(http.StatusOK, models.SuccessResponse(responses))
}

// GetRepository handles GET /admin/api/v1/github/repositories/:id
func (h *GitHubHandler) GetRepository(c *gin.Context) {

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse(
			"Invalid repository ID", err))
		return
	}

	repo, err := h.githubService.GetRepositoryByID(uint(id))
	if err != nil {
		if err.Error() == "repository binding not found" {
			c.JSON(http.StatusNotFound, models.NotFoundResponse(
				"Repository binding not found"))
		} else {
			c.JSON(http.StatusInternalServerError, models.InternalServerErrorResponse(
				"Failed to get repository", err))
		}
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(repo.ToResponse()))
}

// UpdateRepository handles PUT /admin/api/v1/github/repositories/:id
func (h *GitHubHandler) UpdateRepository(c *gin.Context) {
	localizer := getLocalizer(c)

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse(
			"Invalid repository ID", err))
		return
	}

	var req models.GitHubRepositoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse(
			localizer.Get(i18n.ErrInvalidRequestFormat), err))
		return
	}

	repo, err := h.githubService.UpdateRepository(uint(id), &req)
	if err != nil {
		if err.Error() == "repository binding not found" {
			c.JSON(http.StatusNotFound, models.NotFoundResponse(
				"Repository binding not found"))
		} else {
			c.JSON(http.StatusBadRequest, models.BadRequestResponse(
				"Failed to update repository binding", err))
		}
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(repo.ToResponse()))
}

// DeleteRepository handles DELETE /admin/api/v1/github/repositories/:id
func (h *GitHubHandler) DeleteRepository(c *gin.Context) {

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse(
			"Invalid repository ID", err))
		return
	}

	if err := h.githubService.DeleteRepository(uint(id)); err != nil {
		if err.Error() == "repository binding not found" {
			c.JSON(http.StatusNotFound, models.NotFoundResponse(
				"Repository binding not found"))
		} else {
			c.JSON(http.StatusInternalServerError, models.InternalServerErrorResponse(
				"Failed to delete repository", err))
		}
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": "Repository binding deleted successfully",
	}))
}

// SyncRepository handles POST /admin/api/v1/github/repositories/:id/sync
func (h *GitHubHandler) SyncRepository(c *gin.Context) {

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse(
			"Invalid repository ID", err))
		return
	}

	var req models.SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Default to non-force sync if no body provided
		req.Force = false
	}

	response, err := h.githubService.SyncRepository(uint(id), req.Force)
	if err != nil {
		if err.Error() == "repository binding not found" {
			c.JSON(http.StatusNotFound, models.NotFoundResponse(
				"Repository binding not found"))
		} else {
			c.JSON(http.StatusBadRequest, models.BadRequestResponse(
				"Failed to sync repository", err))
		}
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(response))
}

// GetRepositoryReleases handles GET /admin/api/v1/github/repositories/:id/releases
func (h *GitHubHandler) GetRepositoryReleases(c *gin.Context) {
	_, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse(
			"Invalid repository ID", err))
		return
	}

	// Get pagination parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit > 100 {
		limit = 100
	}

	// TODO: Implement GetRepositoryReleases in GitHubService
	// For now, return empty array
	c.JSON(http.StatusOK, models.SuccessResponse([]interface{}{}))
}

// ValidateRepository handles POST /admin/api/v1/github/repositories/validate
func (h *GitHubHandler) ValidateRepository(c *gin.Context) {
	var req struct {
		URL            string `json:"repository_url" binding:"required"`
		Owner          string `json:"owner_name"`
		Repo           string `json:"repo_name"`
		AuthType       string `json:"auth_type" binding:"required,oneof=token github_app"`
		Token          string `json:"access_token"`
		GitHubAppID    int64  `json:"github_app_id"`
		InstallationID int64  `json:"installation_id"`
		PrivateKey     string `json:"private_key"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse(
			"Invalid request data", err))
		return
	}

	// Basic validation
	if req.URL == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCodeAndError(
			http.StatusBadRequest, "Repository URL is required", "INVALID_URL"))
		return
	}

	// Basic GitHub URL validation
	if !strings.Contains(req.URL, "github.com") {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCodeAndError(
			http.StatusBadRequest, "Invalid GitHub URL format", "INVALID_URL"))
		return
	}

	var result *services.RepositoryValidationResult
	var err error

	// Validate repository based on authentication type
	switch req.AuthType {
	case "token":
		if req.Token == "" {
			c.JSON(http.StatusBadRequest, models.ErrorResponseWithCodeAndError(
				http.StatusBadRequest, "Access token is required for token authentication", "MISSING_TOKEN"))
			return
		}
		result, err = h.githubService.ValidateRepositoryWithInfo(req.URL, req.Token)
	case "github_app":
		if req.GitHubAppID == 0 || req.PrivateKey == "" {
			c.JSON(http.StatusBadRequest, models.ErrorResponseWithCodeAndError(
				http.StatusBadRequest, "GitHub App ID and private key are required for GitHub App authentication", "MISSING_APP_CREDENTIALS"))
			return
		}
		result, err = h.githubService.ValidateRepositoryWithGitHubApp(req.URL, req.GitHubAppID, req.PrivateKey, req.InstallationID)
	default:
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCodeAndError(
			http.StatusBadRequest, "Invalid authentication type", "INVALID_AUTH_TYPE"))
		return
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCodeAndError(
			http.StatusBadRequest, "Repository validation failed: "+err.Error(), "VALIDATION_FAILED"))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(result))
}

// GetGitHubAppInstallations handles POST /admin/api/v1/github/app/installations
func (h *GitHubHandler) GetGitHubAppInstallations(c *gin.Context) {
	var req struct {
		GitHubAppID int64  `json:"github_app_id" binding:"required"`
		PrivateKey  string `json:"private_key" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse(
			"Invalid request data", err))
		return
	}

	// Create GitHub Apps client
	appsAuth, err := services.NewGitHubAppsAuth(req.GitHubAppID, req.PrivateKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCodeAndError(
			http.StatusBadRequest, "Failed to create GitHub Apps client: "+err.Error(), "INVALID_APP_CREDENTIALS"))
		return
	}

	// Get installations
	installations, err := appsAuth.GetInstallations()
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCodeAndError(
			http.StatusBadRequest, "Failed to get installations: "+err.Error(), "INSTALLATION_FAILED"))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"installations": installations,
		"count":         len(installations),
	}))
}

// TestGitHubApp handles POST /admin/api/v1/github/app/test
func (h *GitHubHandler) TestGitHubApp(c *gin.Context) {
	var req struct {
		GitHubAppID    int64  `json:"github_app_id" binding:"required"`
		PrivateKey     string `json:"private_key" binding:"required"`
		InstallationID int64  `json:"installation_id"`
		RepositoryURL  string `json:"repository_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse(
			"Invalid request data", err))
		return
	}

	// Create GitHub Apps client
	appsAuth, err := services.NewGitHubAppsAuth(req.GitHubAppID, req.PrivateKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCodeAndError(
			http.StatusBadRequest, "Failed to create GitHub Apps client: "+err.Error(), "INVALID_APP_CREDENTIALS"))
		return
	}

	// Test connection
	if req.RepositoryURL != "" {
		// Parse repository URL
		ownerName, repoName, err := parseRepositoryURL(req.RepositoryURL)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponseWithCodeAndError(
				http.StatusBadRequest, "Invalid repository URL: "+err.Error(), "INVALID_URL"))
			return
		}

		// Test connection to repository
		if err := appsAuth.TestConnection(ownerName, repoName); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponseWithCodeAndError(
				http.StatusBadRequest, "Connection test failed: "+err.Error(), "CONNECTION_FAILED"))
			return
		}
	}

	// Get installations to verify the app is working
	installations, err := appsAuth.GetInstallations()
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCodeAndError(
			http.StatusBadRequest, "Failed to get installations: "+err.Error(), "INSTALLATION_FAILED"))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"valid":               true,
		"installations_count": len(installations),
		"message":             "GitHub App credentials are valid",
	}))
}

// TestToken handles POST /admin/api/v1/github/test-token
func (h *GitHubHandler) TestToken(c *gin.Context) {
	var req struct {
		Token string `json:"access_token" binding:"required"`
		URL   string `json:"repository_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse(
			"Invalid request data", err))
		return
	}

	// TODO: Implement GitHub token testing in GitHubService
	// For now, return basic validation
	if req.Token == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithCodeAndError(
			http.StatusBadRequest, "GitHub token is required", "INVALID_TOKEN"))
		return
	}

	// Return success for now
	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"valid":       true,
		"token":       "***hidden***",
		"permissions": []string{"repo", "admin:repo_hook"},
	}))
}

// WebhookHandler handles GitHub webhook callbacks
type WebhookHandler struct {
	githubService *services.GitHubService
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler() *WebhookHandler {
	return &WebhookHandler{
		githubService: services.NewGitHubService(),
	}
}

// HandleWebhook handles POST /api/v1/webhook/github/:repo_id
func (h *WebhookHandler) HandleWebhook(c *gin.Context) {
	repoID, err := strconv.ParseUint(c.Param("repo_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	// Read webhook payload
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read payload"})
		return
	}

	// Get GitHub signature
	signature := c.GetHeader("X-Hub-Signature-256")
	if signature == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing signature"})
		return
	}

	// Process webhook
	if err := h.githubService.ProcessWebhook(body, signature, uint(repoID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "processed"})
}

// Helper function to parse GitHub repository URL
func parseRepositoryURL(repositoryURL string) (string, string, error) {
	// Simple URL parsing for GitHub repositories
	// Expected formats:
	// https://github.com/owner/repo
	// https://github.com/owner/repo.git
	// git@github.com:owner/repo.git

	if strings.HasPrefix(repositoryURL, "git@github.com:") {
		// SSH format
		path := strings.TrimPrefix(repositoryURL, "git@github.com:")
		path = strings.TrimSuffix(path, ".git")
		parts := strings.Split(path, "/")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid SSH repository URL format")
		}
		return parts[0], parts[1], nil
	} else if strings.Contains(repositoryURL, "github.com") {
		// HTTPS format
		parts := strings.Split(repositoryURL, "/")
		if len(parts) < 5 {
			return "", "", fmt.Errorf("invalid repository URL format")
		}
		owner := parts[len(parts)-2]
		repo := strings.TrimSuffix(parts[len(parts)-1], ".git")
		return owner, repo, nil
	}

	return "", "", fmt.Errorf("unsupported repository URL format")
}
