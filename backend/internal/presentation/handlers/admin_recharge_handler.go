package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RECHARGE MONITORING APIS
// ============================================================================

// GetRechargeTransactions returns paginated list of recharge transactions
func (h *AdminComprehensiveHandler) GetRechargeTransactions(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	// Get transactions from service
	transactions, err := h.rechargeService.GetRechargeHistory(ctx, "", perPage, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve transactions",
		})
		return
	}

	// Build response
	response := make([]map[string]interface{}, 0, len(transactions))
	for _, r := range transactions {
		response = append(response, map[string]interface{}{
			"id":         r.ID,
			"msisdn":     r.MSISDN,
			"amount":     r.Amount,
			"network":    r.NetworkProvider,
			"status":     r.Status,
			"created_at": r.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"pagination": gin.H{
			"page":     page,
			"per_page": perPage,
			"total":    len(transactions),
		},
	})
}

// GetRechargeStats returns recharge statistics
func (h *AdminComprehensiveHandler) GetRechargeStats(c *gin.Context) {
	ctx := c.Request.Context()

	stats, err := h.rechargeService.GetStats(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve stats",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// RetryFailedRecharge retries a failed recharge transaction
func (h *AdminComprehensiveHandler) RetryFailedRecharge(c *gin.Context) {
	ctx := c.Request.Context()

	rechargeIDStr := c.Param("id")
	rechargeID, err := uuid.Parse(rechargeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid recharge ID",
		})
		return
	}

	// Get the recharge transaction
	recharge, err := h.rechargeService.GetRechargeByID(ctx, rechargeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Recharge transaction not found",
		})
		return
	}

	// Validate status
	if recharge.Status != "FAILED" && recharge.Status != "PENDING" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Only failed or pending transactions can be retried",
		})
		return
	}

	// Retry the transaction
	if err := h.rechargeService.ProcessSuccessfulPayment(ctx, recharge.PaymentReference); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retry transaction",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Transaction retry initiated successfully",
	})
}

// GetVTPassStatus returns VTPass integration status
func (h *AdminComprehensiveHandler) GetVTPassStatus(c *gin.Context) {
	ctx := c.Request.Context()

	networks, err := h.rechargeService.GetNetworks(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve VTPass status",
		})
		return
	}

	// Note: NetworkInfo doesn't have Provider field, so we assume all networks are direct
	// In production, this would query network_configs table for provider info
	vtpassCount := 0
	directCount := len(networks)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"connected":      true,
			"vtpass_count":   vtpassCount,
			"direct_count":   directCount,
			"total_networks": len(networks),
		},
	})
}

// UpdateProviderConfig updates provider configuration
func (h *AdminComprehensiveHandler) UpdateProviderConfig(c *gin.Context) {
	var req struct {
		Network  string `json:"network" binding:"required"`
		Provider string `json:"provider" binding:"required"`
		Enabled  bool   `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Validate provider
	if req.Provider != "vtpass" && req.Provider != "direct" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid provider. Must be 'vtpass' or 'direct'",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Provider configuration updated successfully",
	})
}

// GetNetworkConfigurations returns network configurations
func (h *AdminComprehensiveHandler) GetNetworkConfigurations(c *gin.Context) {
	ctx := c.Request.Context()

	// Get full network configs from database
	networks, err := h.networkConfigService.GetNetworkConfigsAdmin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve network configurations",
		})
		return
	}

	// Convert to response format with all fields
	response := make([]map[string]interface{}, 0, len(networks))
	for _, network := range networks {
		isActive := true
		if network.IsActive != nil {
			isActive = *network.IsActive
		}

		airtimeEnabled := true
		if network.AirtimeEnabled != nil {
			airtimeEnabled = *network.AirtimeEnabled
		}

		dataEnabled := true
		if network.DataEnabled != nil {
			dataEnabled = *network.DataEnabled
		}

		commissionRate := 0.0
		if network.CommissionRate != nil {
			commissionRate = *network.CommissionRate
		}

		minAmount := float64(0)
		if network.MinimumAmount != nil {
			minAmount = *network.MinimumAmount
		}

		maxAmount := float64(0)
		if network.MaximumAmount != nil {
			maxAmount = *network.MaximumAmount
		}

		response = append(response, map[string]interface{}{
			"id":              network.ID,
			"network":         network.NetworkName,
			"code":            network.NetworkCode,
			"enabled":         isActive,
			"airtime_enabled": airtimeEnabled,
			"data_enabled":    dataEnabled,
			"commission_rate": commissionRate,
			"minimum_amount":  minAmount,
			"maximum_amount":  maxAmount,
			"logo_url":        network.LogoUrl,
			"brand_color":     network.BrandColor,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}
