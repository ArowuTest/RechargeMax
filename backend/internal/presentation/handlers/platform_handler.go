package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"rechargemax/internal/application/services"
)

// PlatformHandler handles platform-wide public endpoints.
type PlatformHandler struct {
	platformSvc *services.PlatformService
}

// NewPlatformHandler creates a new PlatformHandler.
func NewPlatformHandler(platformSvc *services.PlatformService) *PlatformHandler {
	return &PlatformHandler{platformSvc: platformSvc}
}

// GetStatistics returns platform-wide statistics.
func (h *PlatformHandler) GetStatistics(c *gin.Context) {
	stats, err := h.platformSvc.GetStatistics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to retrieve platform statistics"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": stats})
}

// GetRecentWinners returns recent winners up to the given limit.
func (h *PlatformHandler) GetRecentWinners(c *gin.Context) {
	limit := 4
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	winners, err := h.platformSvc.GetRecentWinners(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch recent winners"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": winners})
}
