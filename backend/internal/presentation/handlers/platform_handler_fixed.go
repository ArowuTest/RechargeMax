package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PlatformHandler handles platform-wide endpoints
type PlatformHandlerFixed struct {
	db *gorm.DB
}

// NewPlatformHandlerFixed creates a new platform handler
func NewPlatformHandlerFixed(db *gorm.DB) *PlatformHandlerFixed {
	return &PlatformHandlerFixed{db: db}
}

// GetStatistics returns platform-wide statistics
func (h *PlatformHandlerFixed) GetStatistics(c *gin.Context) {
	// Use correct table names with timestamp suffix
	const (
		usersTable       = "users_2026_01_30_14_00"
		transactionsTable = "transactions_2026_01_30_14_00"
		drawWinnersTable = "draw_winners_2026_01_30_14_00"
		drawsTable       = "draws_2026_01_30_14_00"
		drawEntriesTable = "draw_entries_2026_01_30_14_00"
	)

	// Get total users count
	var totalUsers int64
	if err := h.db.Table(usersTable).Where("is_active = ?", true).Count(&totalUsers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to count users"})
		return
	}

	// Get total transactions count
	var totalTransactions int64
	if err := h.db.Table(transactionsTable).Where("status = ?", "COMPLETED").Count(&totalTransactions).Error; err != nil {
		// If no completed transactions, try without status filter
		if err := h.db.Table(transactionsTable).Count(&totalTransactions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to count transactions"})
			return
		}
	}

	// Get total prizes count
	var totalPrizes int64
	if err := h.db.Table(drawWinnersTable).Count(&totalPrizes).Error; err != nil {
		// If table doesn't exist or is empty, set to 0
		totalPrizes = 0
	}

	// Get active draw
	var activeDraw struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		PrizePool int64     `json:"prize_pool"`
		EndTime   time.Time `json:"end_time"`
	}
	
	err := h.db.Table(drawsTable).
		Where("status = ? AND end_time > ?", "ACTIVE", time.Now()).
		Order("created_at DESC").
		First(&activeDraw).Error

	var activeDrawData map[string]interface{}
	if err == nil {
		// Count entries for active draw
		var entries int64
		h.db.Table(drawEntriesTable).Where("draw_id = ?", activeDraw.ID).Count(&entries)

		activeDrawData = map[string]interface{}{
			"name":      activeDraw.Name,
			"prizePool": activeDraw.PrizePool,
			"endTime":   activeDraw.EndTime,
			"entries":   entries,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"totalUsers":        totalUsers,
			"totalTransactions": totalTransactions,
			"totalPrizes":       totalPrizes,
			"activeDraw":        activeDrawData,
		},
	})
}

// GetRecentWinners returns recent winners
func (h *PlatformHandlerFixed) GetRecentWinners(c *gin.Context) {
	const (
		usersTable       = "users_2026_01_30_14_00"
		drawWinnersTable = "draw_winners_2026_01_30_14_00"
	)

	limit := 4
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	var winners []struct {
		FullName         string    `json:"full_name"`
		PrizeDescription string    `json:"prize_description"`
		PrizeValue       float64   `json:"prize_value"`
		CreatedAt        time.Time `json:"created_at"`
		NetworkProvider  string    `json:"network_provider"`
		Position         int       `json:"position"`
	}

	err := h.db.Table(drawWinnersTable).
		Select("users.full_name, CONCAT('Position ', draw_winners_2026_01_30_14_00.position, ' Prize') as prize_description, draw_winners_2026_01_30_14_00.prize_amount as prize_value, draw_winners_2026_01_30_14_00.created_at, users.msisdn as network_provider, draw_winners_2026_01_30_14_00.position").
		Joins("LEFT JOIN "+usersTable+" as users ON draw_winners_2026_01_30_14_00.user_id = users.id").
		Where("draw_winners_2026_01_30_14_00.claim_status IN (?)", []string{"CLAIMED", "PROCESSING"}).
		Order("draw_winners_2026_01_30_14_00.created_at DESC").
		Limit(limit).
		Scan(&winners).Error

	if err != nil {
		// If no winners found, return empty array instead of error
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    winners,
	})
}
