package admin

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Run-Panel/VerTree/internal/models"
	"github.com/Run-Panel/VerTree/internal/services"
	"github.com/gin-gonic/gin"
)

// VersionHandler handles admin version management endpoints
type VersionHandler struct {
	versionService *services.VersionService
}

// NewVersionHandler creates a new version handler
func NewVersionHandler() *VersionHandler {
	return &VersionHandler{
		versionService: services.NewVersionService(),
	}
}

// CreateVersion handles POST /admin/api/v1/versions
func (h *VersionHandler) CreateVersion(c *gin.Context) {
	var req models.VersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid request format", err))
		return
	}

	version, err := h.versionService.CreateVersion(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Failed to create version", err))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse(version.ToResponse()))
}

// GetVersions handles GET /admin/api/v1/versions
func (h *VersionHandler) GetVersions(c *gin.Context) {
	// Parse query parameters
	channel := c.Query("channel")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	versions, total, err := h.versionService.ListVersions(channel, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.InternalServerErrorResponse("Failed to get versions", err))
		return
	}

	// Convert to response format
	var versionResponses []*models.VersionResponse
	for _, version := range versions {
		versionResponses = append(versionResponses, version.ToResponse())
	}

	// Calculate pagination
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

	response := models.PaginatedResponse{
		Code:       200,
		Message:    "success",
		Data:       versionResponses,
		Pagination: pagination,
	}

	c.JSON(http.StatusOK, response)
}

// GetVersion handles GET /admin/api/v1/versions/:id
func (h *VersionHandler) GetVersion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid version ID", err))
		return
	}

	version, err := h.versionService.GetVersionByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.NotFoundResponse("Version not found"))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(version.ToResponse()))
}

// UpdateVersion handles PUT /admin/api/v1/versions/:id
func (h *VersionHandler) UpdateVersion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid version ID", err))
		return
	}

	var req models.VersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid request format", err))
		return
	}

	version, err := h.versionService.UpdateVersion(uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Failed to update version", err))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(version.ToResponse()))
}

// PublishVersion handles POST /admin/api/v1/versions/:id/publish
func (h *VersionHandler) PublishVersion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid version ID", err))
		return
	}

	version, err := h.versionService.PublishVersion(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Failed to publish version", err))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(version.ToResponse()))
}

// UnpublishVersion handles POST /admin/api/v1/versions/:id/unpublish
func (h *VersionHandler) UnpublishVersion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid version ID", err))
		return
	}

	version, err := h.versionService.UnpublishVersion(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Failed to unpublish version", err))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(version.ToResponse()))
}

// DeleteVersion handles DELETE /admin/api/v1/versions/:id
func (h *VersionHandler) DeleteVersion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid version ID", err))
		return
	}

	if err := h.versionService.DeleteVersion(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Failed to delete version", err))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]string{"message": "Version deleted successfully"}))
}

// CreateVersionWithUpload handles POST /admin/api/v1/applications/:id/versions/upload
func (h *VersionHandler) CreateVersionWithUpload(c *gin.Context) {
	// Get app_id from URL path parameter
	appID := c.Param("id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Application ID is required", nil))
		return
	}

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(500 << 20); err != nil { // 500MB max
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Failed to parse multipart form", err))
		return
	}

	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("No file uploaded", err))
		return
	}
	defer file.Close()

	// Validate file type
	if !isValidFileType(header.Filename) {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid file type", nil))
		return
	}

	// Create upload directory if it doesn't exist
	uploadDir := "./uploads/versions"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, models.InternalServerErrorResponse("Failed to create upload directory", err))
		return
	}

	// Generate unique filename
	filename := generateUniqueFilename(header.Filename)
	filepath := filepath.Join(uploadDir, filename)

	// Save file
	dst, err := os.Create(filepath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.InternalServerErrorResponse("Failed to create file", err))
		return
	}
	defer dst.Close()

	// Copy file content and calculate size & checksum
	hasher := sha256.New()
	teeReader := io.TeeReader(file, hasher)

	size, err := io.Copy(dst, teeReader)
	if err != nil {
		os.Remove(filepath) // Clean up on error
		c.JSON(http.StatusInternalServerError, models.InternalServerErrorResponse("Failed to save file", err))
		return
	}

	// Calculate checksum
	checksum := fmt.Sprintf("sha256:%x", hasher.Sum(nil))

	// Build version request with form data and URL parameter
	req := models.VersionRequest{
		AppID:             appID, // Get from URL parameter instead of form data
		Version:           c.PostForm("version"),
		Channel:           c.PostForm("channel"),
		Title:             c.PostForm("title"),
		Description:       c.PostForm("description"),
		ReleaseNotes:      c.PostForm("release_notes"),
		BreakingChanges:   c.PostForm("breaking_changes"),
		MinUpgradeVersion: c.PostForm("min_upgrade_version"),
		FileURL:           fmt.Sprintf("/uploads/versions/%s", filename),
		FileSize:          size,
		FileChecksum:      checksum,
		IsForced:          c.PostForm("is_forced") == "true",
	}

	// Validate required fields (app_id is already validated from URL)
	if req.Version == "" || req.Channel == "" || req.Title == "" {
		os.Remove(filepath) // Clean up on error
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Missing required fields: version, channel, title", nil))
		return
	}

	// Create version
	version, err := h.versionService.CreateVersion(&req)
	if err != nil {
		os.Remove(filepath) // Clean up on error
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Failed to create version", err))
		return
	}

	// Check if should publish immediately
	if c.PostForm("publish") == "true" {
		if _, err := h.versionService.PublishVersion(version.ID); err != nil {
			c.JSON(http.StatusBadRequest, models.BadRequestResponse("Version created but failed to publish", err))
			return
		}
	}

	c.JSON(http.StatusCreated, models.SuccessResponse(version.ToResponse()))
}

// UpdateVersionWithUpload handles PUT /admin/api/v1/applications/:id/versions/:version_id/upload
func (h *VersionHandler) UpdateVersionWithUpload(c *gin.Context) {
	// Get app_id from URL path parameter
	appID := c.Param("id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Application ID is required", nil))
		return
	}

	idStr := c.Param("version_id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid version ID", err))
		return
	}

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(500 << 20); err != nil { // 500MB max
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Failed to parse multipart form", err))
		return
	}

	// Get existing version
	existingVersion, err := h.versionService.GetVersionByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.NotFoundResponse("Version not found"))
		return
	}

	// Verify version belongs to the specified application
	if existingVersion.AppID != appID {
		c.JSON(http.StatusForbidden, models.ForbiddenResponse("Version does not belong to specified application"))
		return
	}

	// Build update request with form data
	req := models.VersionRequest{
		AppID:             appID, // Use app_id from URL parameter
		Version:           getFormValueOrDefault(c, "version", existingVersion.Version),
		Channel:           getFormValueOrDefault(c, "channel", existingVersion.Channel),
		Title:             getFormValueOrDefault(c, "title", existingVersion.Title),
		Description:       getFormValueOrDefault(c, "description", existingVersion.Description),
		ReleaseNotes:      getFormValueOrDefault(c, "release_notes", existingVersion.ReleaseNotes),
		BreakingChanges:   getFormValueOrDefault(c, "breaking_changes", existingVersion.BreakingChanges),
		MinUpgradeVersion: getFormValueOrDefault(c, "min_upgrade_version", existingVersion.MinUpgradeVersion),
		FileURL:           existingVersion.FileURL,
		FileSize:          existingVersion.FileSize,
		FileChecksum:      existingVersion.FileChecksum,
		IsForced:          c.PostForm("is_forced") == "true",
	}

	// Check if new file is uploaded
	if file, header, err := c.Request.FormFile("file"); err == nil {
		defer file.Close()

		// Validate file type
		if !isValidFileType(header.Filename) {
			c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid file type", nil))
			return
		}

		// Create upload directory if it doesn't exist
		uploadDir := "./uploads/versions"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, models.InternalServerErrorResponse("Failed to create upload directory", err))
			return
		}

		// Generate unique filename
		filename := generateUniqueFilename(header.Filename)
		filepath := filepath.Join(uploadDir, filename)

		// Save file
		dst, err := os.Create(filepath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.InternalServerErrorResponse("Failed to create file", err))
			return
		}
		defer dst.Close()

		// Copy file content and calculate size & checksum
		hasher := sha256.New()
		teeReader := io.TeeReader(file, hasher)

		size, err := io.Copy(dst, teeReader)
		if err != nil {
			os.Remove(filepath) // Clean up on error
			c.JSON(http.StatusInternalServerError, models.InternalServerErrorResponse("Failed to save file", err))
			return
		}

		// Calculate checksum
		checksum := fmt.Sprintf("sha256:%x", hasher.Sum(nil))

		// Update file info
		req.FileURL = fmt.Sprintf("/uploads/versions/%s", filename)
		req.FileSize = size
		req.FileChecksum = checksum

		// Remove old file if different
		if existingVersion.FileURL != req.FileURL && strings.HasPrefix(existingVersion.FileURL, "/uploads/") {
			oldPath := "." + existingVersion.FileURL
			os.Remove(oldPath)
		}
	}

	// Update version
	version, err := h.versionService.UpdateVersion(uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Failed to update version", err))
		return
	}

	// Check if should publish immediately
	if c.PostForm("publish") == "true" {
		if _, err := h.versionService.PublishVersion(version.ID); err != nil {
			c.JSON(http.StatusBadRequest, models.BadRequestResponse("Version updated but failed to publish", err))
			return
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(version.ToResponse()))
}

// Helper functions

func isValidFileType(filename string) bool {
	validExts := []string{".zip", ".exe", ".dmg", ".pkg", ".deb", ".rpm", ".tar.gz", ".msi"}
	lowerFilename := strings.ToLower(filename)

	for _, ext := range validExts {
		if strings.HasSuffix(lowerFilename, ext) {
			return true
		}
	}
	return false
}

func generateUniqueFilename(originalName string) string {
	ext := filepath.Ext(originalName)
	base := strings.TrimSuffix(originalName, ext)

	// Add timestamp for uniqueness
	return fmt.Sprintf("%s_%d%s", base, currentTimeMillis(), ext)
}

func currentTimeMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func getFormValueOrDefault(c *gin.Context, key, defaultValue string) string {
	if value := c.PostForm(key); value != "" {
		return value
	}
	return defaultValue
}

// CreateVersionWithUploadGlobal creates a new version with file upload (global endpoint)
func (h *VersionHandler) CreateVersionWithUploadGlobal(c *gin.Context) {
	// Get app_id from form data since it's not in the URL
	appID := c.PostForm("app_id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app_id is required in form data"})
		return
	}

	// Temporarily set the app_id parameter for the existing handler
	c.Params = append(c.Params, gin.Param{Key: "id", Value: appID})

	// Call the existing handler
	h.CreateVersionWithUpload(c)
}

// UpdateVersionWithUploadGlobal updates an existing version with file upload (global endpoint)
func (h *VersionHandler) UpdateVersionWithUploadGlobal(c *gin.Context) {
	// Get app_id from form data since it's not in the URL
	appID := c.PostForm("app_id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app_id is required in form data"})
		return
	}

	// Get version_id from URL
	versionID := c.Param("id")

	// Temporarily modify params to match the existing handler expectations
	originalParams := c.Params
	c.Params = gin.Params{
		{Key: "id", Value: appID},
		{Key: "version_id", Value: versionID},
	}

	// Call the existing handler
	h.UpdateVersionWithUpload(c)

	// Restore original params (good practice)
	c.Params = originalParams
}
