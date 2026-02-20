package handlers

import (
"net/http"
"strconv"
"time"

"github.com/gin-gonic/gin"
"gorm.io/gorm"

"rechargemax/internal/domain/entities"
)

// PlatformHandler handles platform-wide endpoints
type PlatformHandler struct {
db *gorm.DB
}

// NewPlatformHandler creates a new platform handler
func NewPlatformHandler(db *gorm.DB) *PlatformHandler {
return &PlatformHandler{db: db}
}

// GetStatistics returns platform-wide statistics
func (h *PlatformHandler) GetStatistics(c *gin.Context) {
// Get total users count
var totalUsers int64
	if err := h.db.Model(&entities.Users{}).Where("is_active = ?", true).Count(&totalUsers).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to count users"})
return
}

// Get total transactions count
var totalTransactions int64
if err := h.db.Model(&entities.Transaction{}).Where("status = ?", "completed").Count(&totalTransactions).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to count transactions"})
return
}

// Get total prizes count
var totalPrizes int64
	if err := h.db.Model(&entities.DrawWinners{}).Count(&totalPrizes).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to count prizes"})
return
}

// Get active draw
var activeDraw entities.Draw
err := h.db.Where("status = ? AND end_time > ?", "ACTIVE", time.Now()).
Order("created_at DESC").
First(&activeDraw).Error

var activeDrawData map[string]interface{}
if err == nil {
// Count entries for active draw
var entries int64
h.db.Model(&entities.DrawEntry{}).Where("draw_id = ?", activeDraw.ID).Count(&entries)

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
"totalUsers":       totalUsers,
"totalTransactions": totalTransactions,
"totalPrizes":      totalPrizes,
"activeDraw":       activeDrawData,
},
})
}

// GetRecentWinners returns recent winners
func (h *PlatformHandler) GetRecentWinners(c *gin.Context) {
limit := 4
if limitStr := c.Query("limit"); limitStr != "" {
if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
limit = parsedLimit
}
}

	var winners []struct {
		FullName          string    `json:"full_name"`
		PrizeDescription  string    `json:"prize_description"`
		PrizeValue        float64   `json:"prize_value"`
		CreatedAt         time.Time `json:"created_at"`
		NetworkProvider   string    `json:"network_provider"`
		Position          int       `json:"position"`
	}

	err := h.db.Table("draw_winners").
		Select("users.full_name, CONCAT('Position ', draw_winners.position, ' Prize') as prize_description, draw_winners.prize_amount as prize_value, draw_winners.created_at, users.msisdn as network_provider, draw_winners.position").
		Joins("LEFT JOIN users ON draw_winners.user_id = users.id").
		Where("draw_winners.claim_status IN (?)", []string{"CLAIMED", "PROCESSING"}).
		Order("draw_winners.created_at DESC").
		Limit(limit).
		Scan(&winners).Error

if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch recent winners"})
return
}

c.JSON(http.StatusOK, gin.H{
"success": true,
"data":    winners,
})
}
