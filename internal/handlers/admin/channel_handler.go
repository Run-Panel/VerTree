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
		c.JSON(http.StatusInternalServerError, models.InternalErrorResponse("Failed to get channels", err))
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
