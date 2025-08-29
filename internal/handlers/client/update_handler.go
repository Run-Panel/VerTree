package client

import (
	"net/http"
	"strconv"

	"github.com/Run-Panel/VerTree/internal/models"
	"github.com/Run-Panel/VerTree/internal/services"
	"github.com/gin-gonic/gin"
)

// UpdateHandler handles client update endpoints
type UpdateHandler struct {
	updateService  *services.UpdateService
	versionService *services.VersionService
}

// NewUpdateHandler creates a new update handler
func NewUpdateHandler() *UpdateHandler {
	return &UpdateHandler{
		updateService:  services.NewUpdateService(),
		versionService: services.NewVersionService(),
	}
}

// CheckUpdate handles POST /api/v1/check-update
func (h *UpdateHandler) CheckUpdate(c *gin.Context) {
	var req models.CheckUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid request format", err))
		return
	}

	// Get client IP
	clientIP := c.ClientIP()
	if forwardedFor := c.GetHeader("X-Forwarded-For"); forwardedFor != "" {
		clientIP = forwardedFor
	}

	response, err := h.updateService.CheckUpdate(&req, clientIP)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Failed to check for updates", err))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(response))
}

// DownloadStarted handles POST /api/v1/download-started
func (h *UpdateHandler) DownloadStarted(c *gin.Context) {
	var req struct {
		Version  string `json:"version" binding:"required"`
		ClientID string `json:"client_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid request format", err))
		return
	}

	// Get client IP
	clientIP := c.ClientIP()
	if forwardedFor := c.GetHeader("X-Forwarded-For"); forwardedFor != "" {
		clientIP = forwardedFor
	}

	if err := h.updateService.RecordDownloadStart(req.Version, req.ClientID, clientIP); err != nil {
		// Don't fail the request if logging fails
		// Just log the error and continue
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]string{"message": "Download started recorded"}))
}

// InstallResult handles POST /api/v1/install-result
func (h *UpdateHandler) InstallResult(c *gin.Context) {
	var req struct {
		Version      string `json:"version" binding:"required"`
		ClientID     string `json:"client_id" binding:"required"`
		Success      bool   `json:"success"`
		ErrorMessage string `json:"error_message"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid request format", err))
		return
	}

	// Get client IP
	clientIP := c.ClientIP()
	if forwardedFor := c.GetHeader("X-Forwarded-For"); forwardedFor != "" {
		clientIP = forwardedFor
	}

	if err := h.updateService.RecordInstallResult(req.Version, req.ClientID, req.Success, req.ErrorMessage, clientIP); err != nil {
		// Don't fail the request if logging fails
		// Just log the error and continue
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]string{"message": "Install result recorded"}))
}

// GetVersions handles GET /api/v1/versions
func (h *UpdateHandler) GetVersions(c *gin.Context) {
	// Get app_id from middleware (set by API key authentication)
	appID, exists := c.Get("app_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.UnauthorizedResponse("Application ID not found"))
		return
	}

	// Parse query parameters
	channel := c.Query("channel")
	limitStr := c.DefaultQuery("limit", "10")
	publishedOnlyStr := c.DefaultQuery("published_only", "true")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	publishedOnly := publishedOnlyStr != "false" // Default to true unless explicitly false

	// Get versions for this app
	versions, err := h.versionService.GetVersionsForApp(appID.(string), channel, limit, publishedOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.InternalServerErrorResponse("Failed to get versions", err))
		return
	}

	// Convert to client response format (simplified, without admin-only fields)
	var versionResponses []map[string]interface{}
	for _, version := range versions {
		versionResponse := map[string]interface{}{
			"version":             version.Version,
			"channel":             version.Channel,
			"title":               version.Title,
			"description":         version.Description,
			"release_notes":       version.ReleaseNotes,
			"download_url":        version.FileURL,
			"file_size":           version.FileSize,
			"file_checksum":       version.FileChecksum,
			"is_forced":           version.IsForced,
			"min_upgrade_version": version.MinUpgradeVersion,
		}

		// Add published_at only if published
		if version.IsPublished && version.PublishTime != nil {
			versionResponse["published_at"] = version.PublishTime
		}

		versionResponses = append(versionResponses, versionResponse)
	}

	c.JSON(http.StatusOK, models.SuccessResponse(versionResponses))
}
