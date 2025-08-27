package client

import (
	"net/http"

	"github.com/Run-Panel/VerTree/internal/models"
	"github.com/Run-Panel/VerTree/internal/services"
	"github.com/gin-gonic/gin"
)

// UpdateHandler handles client update endpoints
type UpdateHandler struct {
	updateService *services.UpdateService
}

// NewUpdateHandler creates a new update handler
func NewUpdateHandler() *UpdateHandler {
	return &UpdateHandler{
		updateService: services.NewUpdateService(),
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
