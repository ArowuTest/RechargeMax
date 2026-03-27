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

// GetPublicWinners returns a paginated winners wall for the public site.
func (h *PlatformHandler) GetPublicWinners(c *gin.Context) {
	page := 1
	limit := 20
	if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 && l <= 100 {
		limit = l
	}
	winners, total, err := h.platformSvc.GetPublicWinners(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch winners"})
		return
	}
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"winners":     winners,
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": totalPages,
		},
	})
}

// GetPublicWinnerByID returns a single winner record for the public detail page.
func (h *PlatformHandler) GetPublicWinnerByID(c *gin.Context) {
	id := c.Param("id")
	winner, err := h.platformSvc.GetPublicWinnerByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Winner not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": winner})
}
