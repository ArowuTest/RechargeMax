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

	// TODO: Implement GetAllRecharges with filters
	_ = network

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    []interface{}{},
	})
}

// GetUSSDStatistics returns USSD recharge statistics
func (h *AdminComprehensiveHandler) GetUSSDStatistics(c *gin.Context) {
	// TODO: Implement statistics aggregation
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_recharges": 0,
			"total_amount":    0,
			"total_points":    0,
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
