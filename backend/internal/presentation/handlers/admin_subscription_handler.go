package handlers

import (
	"fmt"
	"net/http"
	"time"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"rechargemax/internal/domain/entities"
)

// GetSubscriptionTiers returns all subscription tiers
func (h *AdminComprehensiveHandler) GetSubscriptionTiers(c *gin.Context) {
	ctx := c.Request.Context()

	tiers, err := h.subscriptionTierService.GetAllTiers(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve subscription tiers",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tiers,
	})
}

// CreateSubscriptionTier creates a new subscription tier
func (h *AdminComprehensiveHandler) CreateSubscriptionTier(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		Name           string `json:"name" binding:"required"`
		Description    string `json:"description"`
		EntriesPerDay  int    `json:"entries_per_day" binding:"required,min=1"`
		BundleQuantity int    `json:"bundle_quantity" binding:"required,min=1"`
		SortOrder      int    `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	tier, err := h.subscriptionTierService.CreateTier(ctx, req.Name, req.Description, req.EntriesPerDay, req.SortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create subscription tier",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    tier,
	})
}

// UpdateSubscriptionTier updates an existing subscription tier
func (h *AdminComprehensiveHandler) UpdateSubscriptionTier(c *gin.Context) {
	ctx := c.Request.Context()

	tierID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid tier ID",
		})
		return
	}

	var req struct {
		Name          *string `json:"name"`
		Description   *string `json:"description"`
		EntriesPerDay *int    `json:"entries_per_day"`
		SortOrder     *int    `json:"sort_order"`
		IsActive      *bool   `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Get existing tier
	tiers, err := h.subscriptionTierService.GetAllTiers(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve tier",
		})
		return
	}

	var existingTier *entities.SubscriptionTier
	for _, t := range tiers {
		if t.ID == tierID {
			existingTier = t
			break
		}
	}

	if existingTier == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Tier not found",
		})
		return
	}

	// Apply updates
	name := existingTier.Name
	if req.Name != nil {
		name = *req.Name
	}

	description := existingTier.Description
	if req.Description != nil {
		description = *req.Description
	}

	entries := existingTier.Entries
	if req.EntriesPerDay != nil {
		entries = *req.EntriesPerDay
	}

	sortOrder := existingTier.SortOrder
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	isActive := existingTier.IsActive
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	tier, err := h.subscriptionTierService.UpdateTier(ctx, tierID, name, description, entries, sortOrder, isActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update subscription tier",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tier,
	})
}

// DeleteSubscriptionTier deletes a subscription tier
func (h *AdminComprehensiveHandler) DeleteSubscriptionTier(c *gin.Context) {
	ctx := c.Request.Context()

	tierID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid tier ID",
		})
		return
	}

	if err := h.subscriptionTierService.DeleteTier(ctx, tierID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete subscription tier",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Subscription tier deleted successfully",
	})
}

// ============================================================================
// SUBSCRIPTION PRICING
// ============================================================================

// GetCurrentPricing returns current subscription pricing
func (h *AdminComprehensiveHandler) GetCurrentPricing(c *gin.Context) {
	ctx := c.Request.Context()

	pricing, err := h.subscriptionTierService.GetCurrentPricing(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve pricing",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    pricing,
	})
}

// GetPricingHistory returns subscription pricing history from audit_logs
func (h *AdminComprehensiveHandler) GetPricingHistory(c *gin.Context) {
	ctx := c.Request.Context()

	type pricingEntry struct {
		ID          string    `json:"id"          gorm:"column:id"`
		Description string    `json:"description" gorm:"column:description"`
		CreatedAt   time.Time `json:"changed_at"  gorm:"column:created_at"`
		AdminID     string    `json:"admin_id"    gorm:"column:admin_id"`
	}
	var entries []pricingEntry
	h.db.WithContext(ctx).
		Table("audit_logs").
		Where("entity_type = ? AND action IN (?,?)", "subscription_pricing", "UPDATE", "CREATE").
		Order("created_at DESC").
		Limit(100).
		Scan(&entries)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": entries})
}

// UpdatePricing updates subscription pricing
func (h *AdminComprehensiveHandler) UpdatePricing(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		PricePerEntry int64  `json:"price_per_entry" binding:"required,min=1"`
		Reason        string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	pricing, err := h.subscriptionTierService.SetPricePerEntry(ctx, req.PricePerEntry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update pricing",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    pricing,
	})
}

// ============================================================================
// DAILY SUBSCRIPTION MONITORING
// ============================================================================

// GetDailySubscriptions returns all daily subscriptions with filters
func (h *AdminComprehensiveHandler) GetDailySubscriptions(c *gin.Context) {
	ctx := c.Request.Context()

	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))

	// Get all subscriptions
	subscriptions, total, err := h.subscriptionService.GetAllSubscriptions(ctx, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve subscriptions: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"data":     subscriptions,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

// GetDailySubscriptionDetails returns details of a specific subscription
func (h *AdminComprehensiveHandler) GetDailySubscriptionDetails(c *gin.Context) {
	subscriptionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid subscription ID",
		})
		return
	}

	// Load subscription directly from DB
	type subRow struct {
		ID          string  `json:"id"           gorm:"column:id"`
		MSISDN      string  `json:"msisdn"       gorm:"column:msisdn"`
		Status      string  `json:"status"       gorm:"column:status"`
		DailyAmount float64 `json:"daily_amount" gorm:"column:daily_amount"`
		CreatedAt   string  `json:"created_at"   gorm:"column:created_at"`
	}
	var sub subRow
	if err := h.db.WithContext(c.Request.Context()).
		Table("daily_subscriptions").
		Where("id = ?", subscriptionID).
		First(&sub).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Subscription not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": sub})
}

// CancelDailySubscription cancels a daily subscription
func (h *AdminComprehensiveHandler) CancelDailySubscription(c *gin.Context) {
	ctx := c.Request.Context()

	subscriptionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid subscription ID",
		})
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := h.subscriptionTierService.CancelSubscription(ctx, subscriptionID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to cancel subscription",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Subscription cancelled successfully",
	})
}

// GetSubscriptionBillings returns billing history from transactions table
func (h *AdminComprehensiveHandler) GetSubscriptionBillings(c *gin.Context) {
	ctx := c.Request.Context()

	page := 1
	limit := 20
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	if page < 1 { page = 1 }
	if limit < 1 || limit > 100 { limit = 20 }
	offset := (page - 1) * limit

	type billing struct {
		ID          string    `json:"id"          gorm:"column:id"`
		MSISDN      string    `json:"msisdn"      gorm:"column:msisdn"`
		Amount      int64     `json:"amount"      gorm:"column:amount"`
		Status      string    `json:"status"      gorm:"column:status"`
		CreatedAt   time.Time `json:"created_at"  gorm:"column:created_at"`
	}
	var billings []billing
	var total int64
	q := h.db.WithContext(ctx).Table("transactions").
		Where("payment_method = ? OR recharge_type = ?", "subscription", "SUBSCRIPTION")
	q.Count(&total)
	q.Order("created_at DESC").Offset(offset).Limit(limit).Scan(&billings)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    billings,
		"meta": gin.H{"page": page, "limit": limit, "total": total},
	})
}
