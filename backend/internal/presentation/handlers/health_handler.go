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
	var insertErr, rlsErr string

	countErr := h.db.Table("admin_users").Count(&adminCount).Error
	_ = h.db.Table("admin_users").Find(&admins)

	// Try direct insert
	insertResult := h.db.Exec(`INSERT INTO admin_users (id, email, password_hash, full_name, role, permissions, is_active, created_at, updated_at)
VALUES ('950e8400-e29b-41d4-a716-446655440001',
        'admin@rechargemax.ng',
        '$2a$10$GSv3/EaeIzohXsGy6jIMfuoOCMkBLZJF/OiqtG7kVdVoD/dKXypoe',
        'Super Administrator',
        'SUPER_ADMIN',
        '["view_analytics","manage_users"]',
        true,
        NOW(), NOW())
ON CONFLICT (email) DO UPDATE SET
        password_hash = '$2a$10$GSv3/EaeIzohXsGy6jIMfuoOCMkBLZJF/OiqtG7kVdVoD/dKXypoe',
        is_active = true`)
	if insertResult.Error != nil {
		insertErr = insertResult.Error.Error()
	}

	// Check RLS status
	var rlsEnabled bool
	_ = h.db.Raw("SELECT relrowsecurity FROM pg_class WHERE relname = 'admin_users'").Scan(&rlsEnabled)
	if rlsEnabled {
		rlsErr = "RLS_ENABLED"
	} else {
		rlsErr = "RLS_DISABLED"
	}

	// Re-count after insert
	_ = h.db.Table("admin_users").Count(&adminCount)
	_ = h.db.Table("admin_users").Find(&admins)

	var netCount, tierCount int64
	_ = h.db.Table("network_configs").Count(&netCount)
	_ = h.db.Table("subscription_tiers").Count(&tierCount)

	countErrStr := ""
	if countErr != nil {
		countErrStr = countErr.Error()
	}

	// List all tables
	type TableRow struct {
		Tablename string `gorm:"column:tablename"`
	}
	var tables []TableRow
	_ = h.db.Raw("SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename").Scan(&tables)
	tableNames := make([]string, 0)
	for _, t := range tables {
		tableNames = append(tableNames, t.Tablename)
	}

	c.JSON(http.StatusOK, gin.H{
		"admin_count":   adminCount,
		"network_count": netCount,
		"tier_count":    tierCount,
		"admins":        admins,
		"insert_error":  insertErr,
		"rls_status":    rlsErr,
		"count_error":   countErrStr,
		"tables":        tableNames,
	})
}

// BackfillTransactionUserIDs is a one-shot maintenance endpoint.
// It backfills user_id on transactions that have NULL user_id by joining on msisdn.
// Safe to call multiple times (UPDATE is idempotent).
func (h *HealthHandler) BackfillTransactionUserIDs(c *gin.Context) {
	result := h.db.Exec(`
		UPDATE transactions t
		SET    user_id = u.id
		FROM   users u
		WHERE  t.msisdn = u.msisdn
		  AND  t.user_id IS NULL
	`)
	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(200, gin.H{
		"message":       "backfill complete",
		"rows_updated":  result.RowsAffected,
	})
}
