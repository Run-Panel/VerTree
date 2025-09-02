package client

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/Run-Panel/VerTree/internal/models"
	"github.com/Run-Panel/VerTree/internal/services"
	"github.com/gin-gonic/gin"
)

// DownloadHandler handles file download endpoints for clients
type DownloadHandler struct {
	versionService *services.VersionService
	fileService    *services.FileService
}

// NewDownloadHandler creates a new download handler
func NewDownloadHandler() *DownloadHandler {
	return &DownloadHandler{
		versionService: services.NewVersionService(),
		fileService:    services.NewFileService(),
	}
}

// DownloadLatestVersion handles GET /api/v1/download/latest/:app_id/:channel
func (h *DownloadHandler) DownloadLatestVersion(c *gin.Context) {
	appID := c.Param("app_id")
	channel := c.Param("channel")

	if appID == "" || channel == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "app_id and channel are required",
		})
		return
	}

	// Get latest version for app and channel
	version, err := h.versionService.GetLatestVersionForApp(appID, channel)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("No published version found for app %s channel %s", appID, channel),
		})
		return
	}

	// Redirect to version download
	h.downloadVersion(c, version)
}

// DownloadSpecificVersion handles GET /api/v1/download/version/:app_id/:version
func (h *DownloadHandler) DownloadSpecificVersion(c *gin.Context) {
	appID := c.Param("app_id")
	versionStr := c.Param("version")

	if appID == "" || versionStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "app_id and version are required",
		})
		return
	}

	// Get version by app_id and version string
	versions, err := h.versionService.GetVersionsForApp(appID, "", 50, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get versions",
		})
		return
	}

	var version *models.Version
	for _, v := range versions {
		if v.Version == versionStr {
			version = v
			break
		}
	}

	if version == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Version %s not found for app %s", versionStr, appID),
		})
		return
	}

	h.downloadVersion(c, version)
}

// DownloadCachedFile handles GET /api/v1/download/cached/:file_id
func (h *DownloadHandler) DownloadCachedFile(c *gin.Context) {
	fileID, err := strconv.ParseUint(c.Param("file_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid file ID",
		})
		return
	}

	// Serve cached file
	if err := h.fileService.ServeCachedFile(uint(fileID), c.Writer, c.Request); err != nil {
		if err.Error() == "cached file not found" || err.Error() == "cached file no longer exists" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "File not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to serve file",
			})
		}
		return
	}
}

// GetVersionInfo handles GET /api/v1/version-info/:app_id/:channel
func (h *DownloadHandler) GetVersionInfo(c *gin.Context) {
	appID := c.Param("app_id")
	channel := c.Param("channel")

	if appID == "" || channel == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "app_id and channel are required",
		})
		return
	}

	// Get latest version for app and channel
	version, err := h.versionService.GetLatestVersionForApp(appID, channel)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("No published version found for app %s channel %s", appID, channel),
		})
		return
	}

	// Return version information (without triggering download)
	response := version.ToResponse()

	// Add download URL
	downloadURL := fmt.Sprintf("/api/v1/download/latest/%s/%s", appID, channel)

	c.JSON(http.StatusOK, gin.H{
		"version":      response,
		"download_url": downloadURL,
	})
}

// GetVersionHistory handles GET /api/v1/version-history/:app_id/:channel
func (h *DownloadHandler) GetVersionHistory(c *gin.Context) {
	appID := c.Param("app_id")
	channel := c.Param("channel")

	if appID == "" || channel == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "app_id and channel are required",
		})
		return
	}

	// Get query parameters
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	// Get versions for app and channel
	versions, err := h.versionService.GetVersionsForApp(appID, channel, limit, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get version history",
		})
		return
	}

	// Convert to response format with download URLs
	var responses []gin.H
	for _, version := range versions {
		response := version.ToResponse()
		downloadURL := fmt.Sprintf("/api/v1/download/version/%s/%s", appID, version.Version)

		responses = append(responses, gin.H{
			"version":      response,
			"download_url": downloadURL,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"versions": responses,
		"total":    len(responses),
	})
}

// CheckUpdate handles POST /api/v1/check-update (existing endpoint)
func (h *DownloadHandler) CheckUpdate(c *gin.Context) {
	var req struct {
		AppID          string `json:"app_id" validate:"required"`
		CurrentVersion string `json:"current_version" validate:"required"`
		Channel        string `json:"channel" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Get latest version for app and channel
	latestVersion, err := h.versionService.GetLatestVersionForApp(req.AppID, req.Channel)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("No published version found for app %s channel %s", req.AppID, req.Channel),
		})
		return
	}

	// Check if update is available
	hasUpdate := latestVersion.Version != req.CurrentVersion

	response := gin.H{
		"has_update":      hasUpdate,
		"current_version": req.CurrentVersion,
		"latest_version":  latestVersion.ToResponse(),
	}

	if hasUpdate {
		response["download_url"] = fmt.Sprintf("/api/v1/download/latest/%s/%s", req.AppID, req.Channel)
	}

	c.JSON(http.StatusOK, response)
}

// Private helper methods

func (h *DownloadHandler) downloadVersion(c *gin.Context, version *models.Version) {
	// Check if it's a cached file URL
	if isCachedFileURL(version.FileURL) {
		// Extract file ID from cached URL and redirect
		fileID := extractFileIDFromURL(version.FileURL)
		if fileID > 0 {
			c.Redirect(http.StatusFound, fmt.Sprintf("/api/v1/download/cached/%d", fileID))
			return
		}
	}

	// For external URLs, redirect directly
	c.Redirect(http.StatusFound, version.FileURL)
}

func isCachedFileURL(url string) bool {
	// Check if URL is a cached file URL
	return len(url) > 20 && (url[:20] == "/api/v1/download/cac" || url[:25] == "/api/v1/download/cached/")
}

func extractFileIDFromURL(url string) uint {
	// Extract file ID from cached URL pattern: /api/v1/download/cached/{id}
	if len(url) < 25 {
		return 0
	}

	idStr := url[25:] // Skip "/api/v1/download/cached/"
	if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
		return uint(id)
	}

	return 0
}
