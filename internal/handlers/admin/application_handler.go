package admin

import (
	"net/http"
	"strconv"

	"github.com/Run-Panel/VerTree/internal/i18n"
	"github.com/Run-Panel/VerTree/internal/models"
	"github.com/Run-Panel/VerTree/internal/services"
	"github.com/gin-gonic/gin"
)

// ApplicationHandler handles application management endpoints
type ApplicationHandler struct {
	appService *services.ApplicationService
}

// getLocalizer gets the localizer based on request language preference
func getLocalizer(c *gin.Context) i18n.Localizer {
	lang := c.GetHeader("Accept-Language")
	if lang == "" {
		lang = c.Query("lang")
	}
	return i18n.NewLocalizer(lang)
}

// NewApplicationHandler creates a new application handler
func NewApplicationHandler() *ApplicationHandler {
	return &ApplicationHandler{
		appService: services.NewApplicationService(),
	}
}

// CreateApplication handles POST /admin/api/v1/applications
func (h *ApplicationHandler) CreateApplication(c *gin.Context) {
	localizer := getLocalizer(c)

	var req models.ApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse(localizer.Get(i18n.ErrInvalidRequestFormat), err))
		return
	}

	// Get admin ID from context (set by auth middleware)
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.UnauthorizedResponse(localizer.Get(i18n.ErrAdminIDNotFound)))
		return
	}

	app, err := h.appService.CreateApplication(&req, adminID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse(localizer.Get(i18n.ErrApplicationCreateFailed), err))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse(app.ToResponse()))
}

// GetApplications handles GET /admin/api/v1/applications
func (h *ApplicationHandler) GetApplications(c *gin.Context) {
	// Parse pagination parameters
	page := 1
	limit := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Get admin ID from context
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.UnauthorizedResponse("Admin ID not found in context"))
		return
	}

	response, err := h.appService.GetApplications(page, limit, adminID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.InternalServerErrorResponse("Failed to fetch applications", err))
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetApplication handles GET /admin/api/v1/applications/:id
func (h *ApplicationHandler) GetApplication(c *gin.Context) {
	appID := c.Param("id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Application ID is required", nil))
		return
	}

	app, err := h.appService.GetApplication(appID)
	if err != nil {
		if err.Error() == "application not found" {
			c.JSON(http.StatusNotFound, models.NotFoundResponse("Application not found"))
		} else {
			c.JSON(http.StatusInternalServerError, models.InternalServerErrorResponse("Failed to fetch application", err))
		}
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(app.ToResponse()))
}

// UpdateApplication handles PUT /admin/api/v1/applications/:id
func (h *ApplicationHandler) UpdateApplication(c *gin.Context) {
	appID := c.Param("id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Application ID is required", nil))
		return
	}

	var req models.ApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid request format", err))
		return
	}

	// Get admin ID from context
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.UnauthorizedResponse("Admin ID not found in context"))
		return
	}

	app, err := h.appService.UpdateApplication(appID, &req, adminID.(uint))
	if err != nil {
		if err.Error() == "application not found" {
			c.JSON(http.StatusNotFound, models.NotFoundResponse("Application not found"))
		} else {
			c.JSON(http.StatusBadRequest, models.BadRequestResponse("Failed to update application", err))
		}
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(app.ToResponse()))
}

// DeleteApplication handles DELETE /admin/api/v1/applications/:id
func (h *ApplicationHandler) DeleteApplication(c *gin.Context) {
	appID := c.Param("id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Application ID is required", nil))
		return
	}

	// Get admin ID from context
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.UnauthorizedResponse("Admin ID not found in context"))
		return
	}

	if err := h.appService.DeleteApplication(appID, adminID.(uint)); err != nil {
		if err.Error() == "application not found" {
			c.JSON(http.StatusNotFound, models.NotFoundResponse("Application not found"))
		} else {
			c.JSON(http.StatusInternalServerError, models.InternalServerErrorResponse("Failed to delete application", err))
		}
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]string{
		"message": "Application deleted successfully",
	}))
}

// CreateApplicationKey handles POST /admin/api/v1/applications/:id/keys
func (h *ApplicationHandler) CreateApplicationKey(c *gin.Context) {
	localizer := getLocalizer(c)

	appID := c.Param("id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse(localizer.Get(i18n.ErrApplicationIDRequired), nil))
		return
	}

	var req models.ApplicationKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse(localizer.Get(i18n.ErrInvalidRequestFormat), err))
		return
	}

	// Get admin ID from context (set by auth middleware)
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.UnauthorizedResponse(localizer.Get(i18n.ErrAdminIDNotFound)))
		return
	}

	key, err := h.appService.CreateApplicationKey(appID, &req, adminID.(uint))
	if err != nil {
		if err.Error() == "application not found" {
			c.JSON(http.StatusNotFound, models.NotFoundResponse(localizer.Get(i18n.ErrApplicationNotFound)))
		} else {
			c.JSON(http.StatusBadRequest, models.BadRequestResponse(localizer.Get(i18n.ErrAPIKeyCreateFailed), err))
		}
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse(key))
}

// GetApplicationKeys handles GET /admin/api/v1/applications/:id/keys
func (h *ApplicationHandler) GetApplicationKeys(c *gin.Context) {
	appID := c.Param("id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Application ID is required", nil))
		return
	}

	// Get admin ID from context
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.UnauthorizedResponse("Admin ID not found in context"))
		return
	}

	keys, err := h.appService.GetApplicationKeys(appID, adminID.(uint))
	if err != nil {
		if err.Error() == "application not found" {
			c.JSON(http.StatusNotFound, models.NotFoundResponse("Application not found"))
		} else {
			c.JSON(http.StatusInternalServerError, models.InternalServerErrorResponse("Failed to fetch application keys", err))
		}
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(keys))
}

// UpdateApplicationKey handles PUT /admin/api/v1/applications/:id/keys/:keyId
func (h *ApplicationHandler) UpdateApplicationKey(c *gin.Context) {
	appID := c.Param("id")
	keyID := c.Param("keyId")

	if appID == "" || keyID == "" {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Application ID and Key ID are required", nil))
		return
	}

	var req models.ApplicationKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid request format", err))
		return
	}

	// Get admin ID from context
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.UnauthorizedResponse("Admin ID not found in context"))
		return
	}

	key, err := h.appService.UpdateApplicationKey(appID, keyID, &req, adminID.(uint))
	if err != nil {
		if err.Error() == "application key not found" {
			c.JSON(http.StatusNotFound, models.NotFoundResponse("Application key not found"))
		} else {
			c.JSON(http.StatusBadRequest, models.BadRequestResponse("Failed to update application key", err))
		}
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(key))
}

// DeleteApplicationKey handles DELETE /admin/api/v1/applications/:id/keys/:keyId
func (h *ApplicationHandler) DeleteApplicationKey(c *gin.Context) {
	appID := c.Param("id")
	keyID := c.Param("keyId")

	if appID == "" || keyID == "" {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Application ID and Key ID are required", nil))
		return
	}

	// Get admin ID from context
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.UnauthorizedResponse("Admin ID not found in context"))
		return
	}

	if err := h.appService.DeleteApplicationKey(appID, keyID, adminID.(uint)); err != nil {
		if err.Error() == "application key not found" {
			c.JSON(http.StatusNotFound, models.NotFoundResponse("Application key not found"))
		} else {
			c.JSON(http.StatusInternalServerError, models.InternalServerErrorResponse("Failed to delete application key", err))
		}
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]string{
		"message": "Application key deleted successfully",
	}))
}
