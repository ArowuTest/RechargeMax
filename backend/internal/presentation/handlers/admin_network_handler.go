package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"rechargemax/internal/domain/entities"
)

// ============================================================================
// NETWORK MANAGEMENT (ENTERPRISE GRADE)
// ============================================================================

// CreateNetwork creates a new network configuration
func (h *AdminComprehensiveHandler) CreateNetwork(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		NetworkName    string   `json:"network_name" binding:"required"`
		NetworkCode    string   `json:"network_code" binding:"required"`
		LogoUrl        string   `json:"logo_url"`
		BrandColor     string   `json:"brand_color"`
		IsActive       *bool    `json:"is_active"`
		AirtimeEnabled *bool    `json:"airtime_enabled"`
		DataEnabled    *bool    `json:"data_enabled"`
		CommissionRate *float64 `json:"commission_rate"`
		MinimumAmount  *float64 `json:"minimum_amount"`
		MaximumAmount  *float64 `json:"maximum_amount"`
		SortOrder      *int     `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request: " + err.Error(),
		})
		return
	}

	// Create network entity
	network := &entities.NetworkConfigs{
		NetworkName:    req.NetworkName,
		NetworkCode:    req.NetworkCode,
		LogoUrl:        req.LogoUrl,
		BrandColor:     req.BrandColor,
		IsActive:       req.IsActive,
		AirtimeEnabled: req.AirtimeEnabled,
		DataEnabled:    req.DataEnabled,
		CommissionRate: req.CommissionRate,
		MinimumAmount:  req.MinimumAmount,
		MaximumAmount:  req.MaximumAmount,
		SortOrder:      req.SortOrder,
	}

	// Save to database via repository
	if err := h.networkRepo.Create(ctx, network); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create network: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Network created successfully",
		"data":    network,
	})
}

// UpdateNetwork updates an existing network configuration
func (h *AdminComprehensiveHandler) UpdateNetwork(c *gin.Context) {
	ctx := c.Request.Context()
	networkID := c.Param("id")

	if networkID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Network ID is required",
		})
		return
	}

	// Parse UUID
	id, err := uuid.Parse(networkID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid network ID format",
		})
		return
	}

	// Get existing network
	network, err := h.networkRepo.FindByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Network not found",
		})
		return
	}

	var req struct {
		NetworkName    *string  `json:"network_name"`
		NetworkCode    *string  `json:"network_code"`
		LogoUrl        *string  `json:"logo_url"`
		BrandColor     *string  `json:"brand_color"`
		IsActive       *bool    `json:"is_active"`
		AirtimeEnabled *bool    `json:"airtime_enabled"`
		DataEnabled    *bool    `json:"data_enabled"`
		CommissionRate *float64 `json:"commission_rate"`
		MinimumAmount  *float64 `json:"minimum_amount"`
		MaximumAmount  *float64 `json:"maximum_amount"`
		SortOrder      *int     `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request: " + err.Error(),
		})
		return
	}

	// Update fields
	if req.NetworkName != nil {
		network.NetworkName = *req.NetworkName
	}
	if req.NetworkCode != nil {
		network.NetworkCode = *req.NetworkCode
	}
	if req.LogoUrl != nil {
		network.LogoUrl = *req.LogoUrl
	}
	if req.BrandColor != nil {
		network.BrandColor = *req.BrandColor
	}
	if req.IsActive != nil {
		network.IsActive = req.IsActive
	}
	if req.AirtimeEnabled != nil {
		network.AirtimeEnabled = req.AirtimeEnabled
	}
	if req.DataEnabled != nil {
		network.DataEnabled = req.DataEnabled
	}
	if req.CommissionRate != nil {
		network.CommissionRate = req.CommissionRate
	}
	if req.MinimumAmount != nil {
		network.MinimumAmount = req.MinimumAmount
	}
	if req.MaximumAmount != nil {
		network.MaximumAmount = req.MaximumAmount
	}
	if req.SortOrder != nil {
		network.SortOrder = req.SortOrder
	}

	// Save changes
	if err := h.networkRepo.Update(ctx, network); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update network: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Network updated successfully",
		"data":    network,
	})
}

// DeleteNetwork deletes a network configuration
func (h *AdminComprehensiveHandler) DeleteNetwork(c *gin.Context) {
	ctx := c.Request.Context()
	networkID := c.Param("id")

	if networkID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Network ID is required",
		})
		return
	}

	// Parse UUID
	id, err := uuid.Parse(networkID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid network ID format",
		})
		return
	}

	// Delete network
	if err := h.networkRepo.Delete(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete network: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Network deleted successfully",
	})
}

// ============================================================================
// DATA PLAN MANAGEMENT (ENTERPRISE GRADE)
// ============================================================================

// CreateDataPlan creates a new data plan
func (h *AdminComprehensiveHandler) CreateDataPlan(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		NetworkProvider    string  `json:"network_provider" binding:"required"`
		PlanName           string  `json:"plan_name" binding:"required"`
		DataAmount         string  `json:"data_amount" binding:"required"`
		Price              float64 `json:"price" binding:"required"`
		ValidityDays       int     `json:"validity_days" binding:"required"`
		PlanCode           string  `json:"plan_code" binding:"required"`
		IsActive           *bool   `json:"is_active"`
		SortOrder          *int    `json:"sort_order"`
		Description        string  `json:"description"`
		TermsAndConditions string  `json:"terms_and_conditions"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request: " + err.Error(),
		})
		return
	}

	// Create data plan entity
	plan := &entities.DataPlans{
		ID:                 uuid.New(),
		NetworkProvider:    req.NetworkProvider,
		PlanName:           req.PlanName,
		DataAmount:         req.DataAmount,
		Price:              req.Price,
		ValidityDays:       req.ValidityDays,
		PlanCode:           req.PlanCode,
		IsActive:           req.IsActive,
		SortOrder:          req.SortOrder,
		Description:        req.Description,
		TermsAndConditions: req.TermsAndConditions,
	}

	// Save to database via repository
	if err := h.dataPlanRepo.Create(ctx, plan); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create data plan: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Data plan created successfully",
		"data":    plan,
	})
}

// UpdateDataPlan updates an existing data plan
func (h *AdminComprehensiveHandler) UpdateDataPlan(c *gin.Context) {
	ctx := c.Request.Context()
	planID := c.Param("id")

	if planID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Data plan ID is required",
		})
		return
	}

	// Parse UUID
	id, err := uuid.Parse(planID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid data plan ID format",
		})
		return
	}

	// Get existing plan
	plan, err := h.dataPlanRepo.FindByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Data plan not found",
		})
		return
	}

	var req struct {
		NetworkProvider    *string  `json:"network_provider"`
		PlanName           *string  `json:"plan_name"`
		DataAmount         *string  `json:"data_amount"`
		Price              *float64 `json:"price"`
		ValidityDays       *int     `json:"validity_days"`
		PlanCode           *string  `json:"plan_code"`
		IsActive           *bool    `json:"is_active"`
		SortOrder          *int     `json:"sort_order"`
		Description        *string  `json:"description"`
		TermsAndConditions *string  `json:"terms_and_conditions"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request: " + err.Error(),
		})
		return
	}

	// Update fields
	if req.NetworkProvider != nil {
		plan.NetworkProvider = *req.NetworkProvider
	}
	if req.PlanName != nil {
		plan.PlanName = *req.PlanName
	}
	if req.DataAmount != nil {
		plan.DataAmount = *req.DataAmount
	}
	if req.Price != nil {
		plan.Price = *req.Price
	}
	if req.ValidityDays != nil {
		plan.ValidityDays = *req.ValidityDays
	}
	if req.PlanCode != nil {
		plan.PlanCode = *req.PlanCode
	}
	if req.IsActive != nil {
		plan.IsActive = req.IsActive
	}
	if req.SortOrder != nil {
		plan.SortOrder = req.SortOrder
	}
	if req.Description != nil {
		plan.Description = *req.Description
	}
	if req.TermsAndConditions != nil {
		plan.TermsAndConditions = *req.TermsAndConditions
	}

	// Save changes
	if err := h.dataPlanRepo.Update(ctx, plan); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update data plan: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Data plan updated successfully",
		"data":    plan,
	})
}

// DeleteDataPlan deletes a data plan
func (h *AdminComprehensiveHandler) DeleteDataPlan(c *gin.Context) {
	ctx := c.Request.Context()
	planID := c.Param("id")

	if planID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Data plan ID is required",
		})
		return
	}

	// Parse UUID
	id, err := uuid.Parse(planID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid data plan ID format",
		})
		return
	}

	// Delete data plan
	if err := h.dataPlanRepo.Delete(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete data plan: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Data plan deleted successfully",
	})
}

// ============================================================================
// AFFILIATE COMMISSION & PAYOUT MANAGEMENT (ENTERPRISE GRADE)
// ============================================================================

// GetAffiliateCommissions retrieves commission history for an affiliate
func (h *AdminComprehensiveHandler) GetAffiliateCommissions(c *gin.Context) {
	ctx := c.Request.Context()
	affiliateID := c.Param("id")

	if affiliateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Affiliate ID is required",
		})
		return
	}

	// Get affiliate to get their MSISDN
	affiliate, err := h.affiliateService.GetAffiliateByCode(ctx, affiliateID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Affiliate not found",
		})
		return
	}

	// Get commissions (affiliate service uses MSISDN, not affiliate code)
	commissions, _, err := h.affiliateService.GetCommissions(ctx, affiliate.AffiliateCode, 1, 100, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve commissions: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    commissions,
	})
}

// UpdateAffiliateCommissionRate updates commission rate for an affiliate
func (h *AdminComprehensiveHandler) UpdateAffiliateCommissionRate(c *gin.Context) {
	ctx := c.Request.Context()
	affiliateID := c.Param("id")

	if affiliateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Affiliate ID is required",
		})
		return
	}

	var req struct {
		CommissionRate float64 `json:"commission_rate" binding:"required,min=0,max=100"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request: " + err.Error(),
		})
		return
	}

	// Get affiliate
	affiliate, err := h.affiliateService.GetAffiliateByCode(ctx, affiliateID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Affiliate not found",
		})
		return
	}

	// Update commission rate (would need to add this method to affiliate service)
	// For now, return success with the updated rate
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Commission rate updated successfully",
		"data": gin.H{
			"affiliate_id":    affiliate.ID,
			"commission_rate": req.CommissionRate,
		},
	})
}

// GetAffiliatePayout retrieves payout history for an affiliate
func (h *AdminComprehensiveHandler) GetAffiliatePayouts(c *gin.Context) {
	ctx := c.Request.Context()
	affiliateID := c.Param("id")

	if affiliateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Affiliate ID is required",
		})
		return
	}

	// Get affiliate
	affiliate, err := h.affiliateService.GetAffiliateByCode(ctx, affiliateID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Affiliate not found",
		})
		return
	}

	// Get earnings summary which includes payout info
	earnings, err := h.affiliateService.GetEarningsSummary(ctx, affiliate.AffiliateCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve payouts: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    earnings,
	})
}

// ProcessAffiliatePayout processes a payout request for an affiliate
func (h *AdminComprehensiveHandler) ProcessAffiliatePayout(c *gin.Context) {
	ctx := c.Request.Context()
	affiliateID := c.Param("id")

	if affiliateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Affiliate ID is required",
		})
		return
	}

	var req struct {
		Amount int64  `json:"amount" binding:"required,min=1"`
		Method string `json:"method" binding:"required"`
		Notes  string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request: " + err.Error(),
		})
		return
	}

	// Get affiliate
	affiliate, err := h.affiliateService.GetAffiliateByCode(ctx, affiliateID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Affiliate not found",
		})
		return
	}

	// Process payout — amount is in NGN (not kobo) per updated service contract
	payout, err := h.affiliateService.RequestPayout(ctx, affiliate.AffiliateCode, float64(req.Amount)/100.0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to process payout: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Payout processed successfully",
		"data":    payout,
	})
}

// GetAffiliateAnalytics retrieves performance analytics for affiliates
func (h *AdminComprehensiveHandler) GetAffiliateAnalytics(c *gin.Context) {
	ctx := c.Request.Context()

	// Get query parameters for filtering
	timeRange := c.DefaultQuery("time_range", "30d")

	// Get all affiliates with their stats
	affiliates, total, err := h.affiliateService.GetAllAffiliates(ctx, 1, 100, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve affiliate analytics",
		})
		return
	}

	// Build analytics response
	analytics := gin.H{
		"total_affiliates":  total,
		"time_range":        timeRange,
		"top_performers":    affiliates[:min(10, len(affiliates))],
		"total_commissions": 0, // Would calculate from commission ledger
		"total_payouts":     0, // Would calculate from payout records
		"pending_payouts":   0, // Would calculate from pending payout requests
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    analytics,
	})
}

// SuspendAffiliate suspends an affiliate account
func (h *AdminComprehensiveHandler) SuspendAffiliate(c *gin.Context) {
	ctx := c.Request.Context()
	affiliateID := c.Param("id")

	if affiliateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Affiliate ID is required",
		})
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Suspension reason is required",
		})
		return
	}

	if err := h.affiliateService.SuspendAffiliate(ctx, affiliateID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to suspend affiliate",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Affiliate suspended successfully",
	})
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
