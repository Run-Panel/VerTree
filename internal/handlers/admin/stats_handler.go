package admin

import (
	"net/http"

	"github.com/Run-Panel/VerTree/internal/models"
	"github.com/Run-Panel/VerTree/internal/services"
	"github.com/gin-gonic/gin"
)

// StatsHandler handles admin statistics endpoints
type StatsHandler struct {
	statsService *services.StatsService
}

// NewStatsHandler creates a new stats handler
func NewStatsHandler() *StatsHandler {
	return &StatsHandler{
		statsService: services.NewStatsService(),
	}
}

// GetStats handles GET /admin/api/v1/stats
func (h *StatsHandler) GetStats(c *gin.Context) {
	var req models.StatsRequest

	// Set defaults
	req.Period = c.DefaultQuery("period", "7d")
	req.Action = c.DefaultQuery("action", "all")

	// Validate query parameters
	if req.Period != "1d" && req.Period != "7d" && req.Period != "30d" && req.Period != "90d" {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid period. Must be one of: 1d, 7d, 30d, 90d", nil))
		return
	}

	if req.Action != "all" && req.Action != "check" && req.Action != "download" && req.Action != "install" && req.Action != "success" && req.Action != "failed" {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid action. Must be one of: all, check, download, install, success, failed", nil))
		return
	}

	stats, err := h.statsService.GetStats(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.InternalServerErrorResponse("Failed to get statistics", err))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(stats))
}

// GetVersionDistribution handles GET /admin/api/v1/stats/distribution
func (h *StatsHandler) GetVersionDistribution(c *gin.Context) {
	// Get period parameter with default
	period := c.DefaultQuery("period", "7d")

	// Validate period parameter
	if period != "1d" && period != "7d" && period != "30d" && period != "90d" {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid period. Must be one of: 1d, 7d, 30d, 90d", nil))
		return
	}

	distribution, err := h.statsService.GetVersionDistribution(period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.InternalServerErrorResponse("Failed to get version distribution", err))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(distribution))
}

// GetRegionDistribution handles GET /admin/api/v1/stats/regions
func (h *StatsHandler) GetRegionDistribution(c *gin.Context) {
	// Get period parameter with default
	period := c.DefaultQuery("period", "7d")

	// Validate period parameter
	if period != "1d" && period != "7d" && period != "30d" && period != "90d" {
		c.JSON(http.StatusBadRequest, models.BadRequestResponse("Invalid period. Must be one of: 1d, 7d, 30d, 90d", nil))
		return
	}

	distribution, err := h.statsService.GetRegionDistribution(period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.InternalServerErrorResponse("Failed to get region distribution", err))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(distribution))
}
