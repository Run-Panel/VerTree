package admin

import (
	"net/http"
	"strconv"

	"github.com/Run-Panel/VerTree/internal/models"
	"github.com/Run-Panel/VerTree/internal/services"
	"github.com/gin-gonic/gin"
)

// ChannelHandler handles admin channel management endpoints
type ChannelHandler struct {
	channelService *services.ChannelService
}

// NewChannelHandler creates a new channel handler
func NewChannelHandler() *ChannelHandler {
	return &ChannelHandler{
		channelService: services.NewChannelService(),
	}
}

// GetChannels handles GET /admin/api/v1/channels
func (h *ChannelHandler) GetChannels(c *gin.Context) {
	channels, err := h.channelService.GetAllChannels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.InternalServerErrorResponse("Failed to get channels", err))
		return
	}

	// Convert to response format
	var channelResponses []*models.ChannelResponse
	for _, channel := range channels {
		channelResponses = append(channelResponses, channel.ToResponse())
	}

	c.JSON(http.StatusOK, models.SuccessResponse(channelResponses))
}

// GetChannel handles GET /admin/api/v1/channels/:id
func (h *ChannelHandler) GetChannel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid channel ID", err))
		return
	}

	channel, err := h.channelService.GetChannelByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.NotFoundResponse("Channel not found"))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(channel.ToResponse()))
}

// CreateChannel handles POST /admin/api/v1/channels
func (h *ChannelHandler) CreateChannel(c *gin.Context) {
	var req models.ChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid request format", err))
		return
	}

	channel, err := h.channelService.CreateChannel(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Failed to create channel", err))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse(channel.ToResponse()))
}

// UpdateChannel handles PUT /admin/api/v1/channels/:id
func (h *ChannelHandler) UpdateChannel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid channel ID", err))
		return
	}

	var req models.ChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid request format", err))
		return
	}

	channel, err := h.channelService.UpdateChannel(uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Failed to update channel", err))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(channel.ToResponse()))
}

// DeleteChannel handles DELETE /admin/api/v1/channels/:id
func (h *ChannelHandler) DeleteChannel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid channel ID", err))
		return
	}

	if err := h.channelService.DeleteChannel(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Failed to delete channel", err))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]string{"message": "Channel deleted successfully"}))
}

// GetChannelsByApp handles GET /admin/api/v1/applications/:id/channels
// Returns enabled channels for the specific application
func (h *ChannelHandler) GetChannelsByApp(c *gin.Context) {
	appID := c.Param("id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("App ID is required", nil))
		return
	}

	// Get application-specific channel configurations
	channels, err := h.channelService.GetChannelsByApp(appID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.InternalServerErrorResponse("Failed to get channels", err))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(channels))
}

// GetAllChannelsForApp handles GET /admin/api/v1/applications/:id/channels/all
// Returns all channels (enabled and disabled) for the specific application
func (h *ChannelHandler) GetAllChannelsForApp(c *gin.Context) {
	appID := c.Param("id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("App ID is required", nil))
		return
	}

	// Get all application-specific channel configurations
	channels, err := h.channelService.GetAllChannelsForApp(appID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.InternalServerErrorResponse("Failed to get channels", err))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(channels))
}

// EnableChannelForApp handles PUT /admin/api/v1/applications/:id/channels/:channel
// Enables/configures a channel for a specific application
func (h *ChannelHandler) EnableChannelForApp(c *gin.Context) {
	appID := c.Param("id")
	channelName := c.Param("channel")

	if appID == "" || channelName == "" {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("App ID and channel name are required", nil))
		return
	}

	var req models.ApplicationChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid request format", err))
		return
	}

	// Set the app ID and channel name from URL params
	req.AppID = appID
	req.ChannelName = channelName

	appChannel, err := h.channelService.EnableChannelForApp(appID, channelName, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Failed to configure channel", err))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(appChannel))
}

// DisableChannelForApp handles DELETE /admin/api/v1/applications/:id/channels/:channel
// Disables a channel for a specific application
func (h *ChannelHandler) DisableChannelForApp(c *gin.Context) {
	appID := c.Param("id")
	channelName := c.Param("channel")

	if appID == "" || channelName == "" {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("App ID and channel name are required", nil))
		return
	}

	if err := h.channelService.DisableChannelForApp(appID, channelName); err != nil {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Failed to disable channel", err))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]string{"message": "Channel disabled successfully"}))
}
