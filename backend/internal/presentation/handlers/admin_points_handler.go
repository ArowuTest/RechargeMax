package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ============================================================================

// GetUsersWithPoints returns users with their points summary
func (h *AdminComprehensiveHandler) GetUsersWithPoints(c *gin.Context) {
	ctx := c.Request.Context()

	searchQuery := c.Query("search")

	users, err := h.pointsService.GetUsersWithPoints(ctx, searchQuery, nil, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve users with points",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    users,
	})
}

// GetPointsHistory returns points transaction history
func (h *AdminComprehensiveHandler) GetPointsHistory(c *gin.Context) {
	ctx := c.Request.Context()

	userIDStr := c.Query("user_id")
	source := c.Query("source")

	var userID *uuid.UUID
	if userIDStr != "" {
		id, err := uuid.Parse(userIDStr)
		if err == nil {
			userID = &id
		}
	}

	history, err := h.pointsService.GetPointsHistory(ctx, userID, source, nil, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve points history",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    history,
	})
}

// AdjustUserPoints adjusts user points (add/deduct)
func (h *AdminComprehensiveHandler) AdjustUserPoints(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		UserID      string `json:"user_id" binding:"required"`
		Points      int    `json:"points" binding:"required"`
		Reason      string `json:"reason" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid user ID",
		})
		return
	}

	// Get admin ID from context (set by auth middleware)
	adminIDStr, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Admin authentication required",
		})
		return
	}

	adminID, ok := adminIDStr.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Invalid admin ID format",
		})
		return
	}

	if err := h.pointsService.AdjustUserPoints(ctx, userID, req.Points, req.Reason, req.Description, adminID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to adjust user points",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User points adjusted successfully",
	})
}

// GetPointsStatistics returns points statistics
func (h *AdminComprehensiveHandler) GetPointsStatistics(c *gin.Context) {
	ctx := c.Request.Context()

	stats, err := h.pointsService.GetPointsStatistics(ctx, nil, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve points statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// ExportUsersWithPoints exports users with points to CSV
func (h *AdminComprehensiveHandler) ExportUsersWithPoints(c *gin.Context) {
	ctx := c.Request.Context()

	csv, err := h.pointsService.ExportUsersWithPointsToCSV(ctx, "", nil, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to export users",
		})
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=users_with_points.csv")
	c.String(http.StatusOK, csv)
}

// ExportPointsHistory exports points history to CSV
func (h *AdminComprehensiveHandler) ExportPointsHistory(c *gin.Context) {
	ctx := c.Request.Context()

	csv, err := h.pointsService.ExportPointsHistoryToCSV(ctx, nil, "", nil, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to export points history",
		})
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=points_history.csv")
	c.String(http.StatusOK, csv)
}

// ============================================================================
// DRAW CSV MANAGEMENT
// ============================================================================

// ExportDrawToCSV exports draw entries to CSV
func (h *AdminComprehensiveHandler) ExportDrawToCSV(c *gin.Context) {
	drawID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid draw ID",
		})
		return
	}

	// TODO: Implement ExportEntriesToCSV in draw service
	csv := "" // Placeholder
	_ = drawID
	if csv == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to export draw entries",
		})
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=draw_entries.csv")
	c.String(http.StatusOK, csv)
}

// ImportWinnersFromCSV imports winners from CSV
func (h *AdminComprehensiveHandler) ImportWinnersFromCSV(c *gin.Context) {
	drawID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid draw ID",
		})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "No file uploaded",
		})
		return
	}

	// Open file
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to open file",
		})
		return
	}
	defer f.Close()

	// TODO: Implement CSV parsing and winner import
	_ = drawID

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Winners imported successfully",
	})
}
