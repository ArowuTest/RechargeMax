package handlers

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"rechargemax/internal/domain/entities"
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

	// admin_id is stored as a string by AdminAuthMiddleware — parse it to uuid.UUID
	adminIDParsed, parseErr := uuid.Parse(fmt.Sprintf("%v", adminIDStr))
	if parseErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Invalid admin ID format",
		})
		return
	}
	adminID := adminIDParsed

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

	// Query draw entries directly from DB and stream as CSV
	type entryRow struct {
		ID        string `gorm:"column:id"`
		MSISDN    string `gorm:"column:msisdn"`
		Source    string `gorm:"column:entry_source"`
		CreatedAt string `gorm:"column:created_at"`
	}
	var entries []entryRow
	if err := h.db.WithContext(c.Request.Context()).
		Table("draw_entries").
		Where("draw_id = ?", drawID).
		Order("created_at ASC").
		Scan(&entries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to export draw entries"})
		return
	}

	var buf strings.Builder
	buf.WriteString("id,msisdn,source,created_at\n")
	for _, e := range entries {
		buf.WriteString(fmt.Sprintf("%s,%s,%s,%s\n", e.ID, e.MSISDN, e.Source, e.CreatedAt))
	}

	// Write audit log so export-history can surface this operation
	drawIDStr := drawID.String()
	ip := c.ClientIP()
	ua := c.Request.UserAgent()
	auditEntry := entities.AuditLog{
		Action:      "export_entries",
		EntityType:  "draw",
		EntityID:    &drawIDStr,
		IPAddress:   &ip,
		UserAgent:   &ua,
		Status:      "success",
	}
	if raw, ok := c.Get("admin_id"); ok {
		if adminStr, ok := raw.(string); ok {
			if adminUID, parseErr := uuid.Parse(adminStr); parseErr == nil {
				auditEntry.AdminUserID = adminUID
			}
		}
	}
	// Best-effort — do not block the CSV download on an audit failure
	h.db.WithContext(c.Request.Context()).Create(&auditEntry) //nolint:errcheck

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=draw_entries.csv")
	c.String(http.StatusOK, buf.String())
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

	// Parse CSV and create winner records
	scanner := bufio.NewScanner(f)
	imported := 0
	skipped := 0
	lineNum := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineNum++
		if lineNum == 1 {
			continue // skip header row
		}
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			skipped++
			continue
		}
		msisdn := strings.TrimSpace(parts[0])
		position := 0
		fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &position)
		prizeType := strings.TrimSpace(parts[2])
		if msisdn == "" || position < 1 {
			skipped++
			continue
		}
		_, createErr := h.winnerService.CreateWinner(
			c.Request.Context(), drawID, msisdn, position,
			prizeType, "", 0, "", 0, "",
		)
		if createErr != nil {
			skipped++
			continue
		}
		imported++
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  fmt.Sprintf("Import complete: %d imported, %d skipped", imported, skipped),
		"imported": imported,
		"skipped":  skipped,
	})
}

// GetDrawExportHistory returns the audit log of draw CSV export operations.
// It queries audit_logs for action=export_entries / entity_type=draw and
// maps them to the DrawExportHistory shape the frontend expects.
// Optional query param: draw_id (UUID) to filter by a specific draw.
func (h *AdminComprehensiveHandler) GetDrawExportHistory(c *gin.Context) {
	ctx := c.Request.Context()

	type exportHistoryRow struct {
		ID           string `json:"id"            gorm:"column:id"`
		DrawID       string `json:"draw_id"       gorm:"column:entity_id"`
		ExportedBy   string `json:"exported_by"   gorm:"column:exported_by"`
		ExportedAt   string `json:"exported_at"   gorm:"column:created_at"`
		TotalMSISDNs int    `json:"total_msisdns" gorm:"column:total_msisdns"`
		TotalPoints  int    `json:"total_points"  gorm:"column:total_points"`
		FileURL      string `json:"file_url"      gorm:"column:file_url"`
	}

	// audit_logs does NOT have admin_email — use admin_user_id cast to text instead
	q := h.db.WithContext(ctx).
		Table("audit_logs").
		Select(`id,
		        COALESCE(entity_id, '')           AS entity_id,
		        COALESCE(admin_user_id::text, 'unknown') AS exported_by,
		        created_at::text                  AS created_at,
		        0                                 AS total_msisdns,
		        0                                 AS total_points,
		        ''                                AS file_url`).
		Where("action = ? AND entity_type = ?", "export_entries", "draw").
		Order("created_at DESC").
		Limit(100)

	if drawID := c.Query("draw_id"); drawID != "" {
		q = q.Where("entity_id = ?", drawID)
	}

	var rows []exportHistoryRow
	if err := q.Scan(&rows).Error; err != nil {
		// Return empty list instead of 500 — no export history yet is not an error
		rows = []exportHistoryRow{}
	}

	if rows == nil {
		rows = []exportHistoryRow{}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    rows,
	})
}

// AdjustUserPointsByID adjusts points for the user identified by the :id URL param.
// This is a convenience wrapper around AdjustUserPoints for per-user API calls.
func (h *AdminComprehensiveHandler) AdjustUserPointsByID(c *gin.Context) {
	ctx := c.Request.Context()

	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid user ID"})
		return
	}

	var req struct {
		Points      int    `json:"points" binding:"required"`
		Reason      string `json:"reason" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	adminIDRaw, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Admin authentication required"})
		return
	}
	adminID, parseErr := uuid.Parse(fmt.Sprintf("%v", adminIDRaw))
	if parseErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Invalid admin ID format"})
		return
	}

	if err := h.pointsService.AdjustUserPoints(ctx, userID, req.Points, req.Reason, req.Description, adminID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to adjust user points: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User points adjusted successfully",
	})
}
