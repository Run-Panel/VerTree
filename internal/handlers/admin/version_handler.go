package admin

import (
	"net/http"
	"strconv"

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
		c.JSON(http.StatusInternalServerError, models.InternalErrorResponse("Failed to get versions", err))
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
