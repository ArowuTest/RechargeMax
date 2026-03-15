package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	// Check database connection
	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "unhealthy",
			"database":  "disconnected",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "unhealthy",
			"database":  "unreachable",
			"error":     err.Error(),
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"database":  "connected",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// DebugDB returns diagnostic info about the database (TEMPORARY - remove before prod)
func (h *HealthHandler) DebugDB(c *gin.Context) {
	type AdminRow struct {
		ID       string `gorm:"column:id"`
		Email    string `gorm:"column:email"`
		IsActive *bool  `gorm:"column:is_active"`
		Role     string `gorm:"column:role"`
	}

	var admins []AdminRow
	var adminCount int64

	_ = h.db.Table("admin_users").Count(&adminCount)
	_ = h.db.Table("admin_users").Find(&admins)

	var netCount, tierCount int64
	_ = h.db.Table("network_configs").Count(&netCount)
	_ = h.db.Table("subscription_tiers").Count(&tierCount)

	c.JSON(http.StatusOK, gin.H{
		"admin_count":      adminCount,
		"network_count":    netCount,
		"tier_count":       tierCount,
		"admins":           admins,
	})
}
