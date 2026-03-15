package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ============================================================================

// GetUSSDRecharges returns USSD recharges with filters
func (h *AdminComprehensiveHandler) GetUSSDRecharges(c *gin.Context) {
	ctx := c.Request.Context()

	msisdn := c.Query("msisdn")
	network := c.Query("network")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate time.Time
	if startDateStr != "" {
		startDate, _ = time.Parse("2006-01-02", startDateStr)
	}
	if endDateStr != "" {
		endDate, _ = time.Parse("2006-01-02", endDateStr)
	}

	if msisdn != "" {
		recharges, err := h.ussdRechargeService.GetUSSDRechargesByMSISDN(ctx, msisdn, startDate, endDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to retrieve USSD recharges",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    recharges,
		})
		return
	}

	// Query ussd_recharges with optional network and date filters
	type ussdRow struct {
		ID              string    `json:"id"               gorm:"column:id"`
		MSISDN          string    `json:"msisdn"           gorm:"column:msisdn"`
		NetworkProvider string    `json:"network_provider" gorm:"column:network_provider"`
		Amount          int64     `json:"amount"           gorm:"column:amount"`
		PointsEarned    int       `json:"points_earned"    gorm:"column:points_earned"`
		Status          string    `json:"status"           gorm:"column:status"`
		CreatedAt       time.Time `json:"created_at"       gorm:"column:created_at"`
	}
	var rows []ussdRow
	q := h.db.WithContext(ctx).Table("ussd_recharges")
	if network != "" {
		q = q.Where("network_provider = ?", network)
	}
	if !startDate.IsZero() {
		q = q.Where("created_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		q = q.Where("created_at <= ?", endDate.Add(24*time.Hour))
	}
	q.Order("created_at DESC").Limit(200).Scan(&rows)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": rows})
}

// GetUSSDStatistics returns USSD recharge aggregate statistics
func (h *AdminComprehensiveHandler) GetUSSDStatistics(c *gin.Context) {
	ctx := c.Request.Context()

	type result struct {
		TotalRecharges int64 `gorm:"column:total_recharges"`
		TotalAmount    int64 `gorm:"column:total_amount"`
		TotalPoints    int64 `gorm:"column:total_points"`
	}
	var res result
	h.db.WithContext(ctx).Raw(`
		SELECT
			COUNT(*)                   AS total_recharges,
			COALESCE(SUM(amount), 0)   AS total_amount,
			COALESCE(SUM(points_earned), 0) AS total_points
		FROM ussd_recharges
	`).Scan(&res)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_recharges": res.TotalRecharges,
			"total_amount":    res.TotalAmount,
			"total_points":    res.TotalPoints,
		},
	})
}

// GetUSSDWebhookLogs returns webhook logs for debugging
func (h *AdminComprehensiveHandler) GetUSSDWebhookLogs(c *gin.Context) {
	ctx := c.Request.Context()

	provider := c.Query("provider")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate time.Time
	if startDateStr != "" {
		startDate, _ = time.Parse("2006-01-02", startDateStr)
	}
	if endDateStr != "" {
		endDate, _ = time.Parse("2006-01-02", endDateStr)
	}

	logs, err := h.ussdRechargeService.GetWebhookLogs(ctx, provider, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve webhook logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
	})
}

// RetryFailedUSSDWebhooks retries failed webhook processing
func (h *AdminComprehensiveHandler) RetryFailedUSSDWebhooks(c *gin.Context) {
	ctx := c.Request.Context()

	if err := h.ussdRechargeService.RetryFailedWebhooks(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retry webhooks",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Failed webhooks retried successfully",
	})
}
