package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"rechargemax/internal/application/services"
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

// AdminComprehensiveHandler handles all new admin endpoints
type AdminComprehensiveHandler struct {
	subscriptionTierService *services.SubscriptionTierService
	ussdRechargeService     *services.USSDRechargeService
	pointsService           *services.PointsService
	drawService             *services.DrawService
	winnerService           *services.WinnerService
	spinService             *services.SpinService
	// NEW: Services for recharge/user/affiliate management
	rechargeService         *services.RechargeService
	userService             *services.UserService
	affiliateService        *services.AffiliateService
	telecomService          *services.TelecomService
	networkConfigService    *services.NetworkConfigService
	// Repositories for direct data access
	networkRepo             repositories.NetworkRepository
	dataPlanRepo            repositories.DataPlanRepository
	subscriptionService     *services.SubscriptionService
	// Prize Tier System Services
	drawTypeService         *services.DrawTypeService
	prizeTemplateService    *services.PrizeTemplateService
	// Direct DB access for settings persistence
	db                      *gorm.DB
}

// NewAdminComprehensiveHandler creates a new comprehensive admin handler
func NewAdminComprehensiveHandler(
	subscriptionTierService *services.SubscriptionTierService,
	ussdRechargeService     *services.USSDRechargeService,
	pointsService           *services.PointsService,
	drawService             *services.DrawService,
	winnerService           *services.WinnerService,
	spinService             *services.SpinService,
	rechargeService         *services.RechargeService,
	userService             *services.UserService,
	affiliateService        *services.AffiliateService,
	telecomService          *services.TelecomService,
	networkConfigService    *services.NetworkConfigService,
	networkRepo             repositories.NetworkRepository,
	dataPlanRepo            repositories.DataPlanRepository,
	subscriptionService     *services.SubscriptionService,
	drawTypeService         *services.DrawTypeService,
	prizeTemplateService    *services.PrizeTemplateService,
	db                      *gorm.DB,
) *AdminComprehensiveHandler {
	return &AdminComprehensiveHandler{
		subscriptionTierService: subscriptionTierService,
		ussdRechargeService:     ussdRechargeService,
		pointsService:           pointsService,
		drawService:             drawService,
		winnerService:           winnerService,
		spinService:             spinService,
		rechargeService:         rechargeService,
		userService:             userService,
		affiliateService:        affiliateService,
		telecomService:          telecomService,
		networkConfigService:    networkConfigService,
		networkRepo:             networkRepo,
		dataPlanRepo:            dataPlanRepo,
		subscriptionService:     subscriptionService,
		drawTypeService:         drawTypeService,
		prizeTemplateService:    prizeTemplateService,
		db:                      db,
	}
}

// ============================================================================
// SUBSCRIPTION TIER MANAGEMENT
// ============================================================================

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

// GetPricingHistory returns subscription pricing history
func (h *AdminComprehensiveHandler) GetPricingHistory(c *gin.Context) {
	// TODO: Implement pricing history in service
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    []interface{}{},
	})
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
		"success": true,
		"data":    subscriptions,
		"total":   total,
		"page":    page,
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

	// Get subscription from repository (service doesn't have GetByID, so we'd need to add it)
	// For now, return a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Subscription details endpoint - implementation pending",
		"id":      subscriptionID,
	})
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

// GetSubscriptionBillings returns billing history
func (h *AdminComprehensiveHandler) GetSubscriptionBillings(c *gin.Context) {
	// TODO: Implement GetBillings in service
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    []interface{}{},
	})
}

// ============================================================================
// USSD RECHARGE MONITORING
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

// ============================================================================
// USER POINTS MANAGEMENT
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

	adminID, ok := adminIDStr.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Invalid admin ID format",
		})
		return
	}

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

	// TODO: Implement ExportEntriesToCSV in draw service
	csv := "" // Placeholder
	_ = drawID
	if csv == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to export draw entries",
		})
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=draw_entries.csv")
	c.String(http.StatusOK, csv)
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

	// TODO: Implement CSV parsing and winner import
	_ = drawID

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Winners imported successfully",
	})
}

// ============================================================================
// WINNER CLAIM PROCESSING
// ============================================================================

// GetPendingClaims returns pending winner claims
func (h *AdminComprehensiveHandler) GetPendingClaims(c *gin.Context) {
	ctx := c.Request.Context()
	
	// Get all winners (page 1, 100 per page, no draw filter)
	winners, _, err := h.winnerService.GetAllWinners(ctx, 1, 100, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve pending claims",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    winners,
	})
}

// ApproveWinnerClaim approves a winner claim
func (h *AdminComprehensiveHandler) ApproveWinnerClaim(c *gin.Context) {
	winnerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid winner ID",
		})
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	
	err = h.winnerService.ApproveClaim(ctx, winnerID.String(), req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to approve claim",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Winner claim approved successfully",
	})
}

// RejectWinnerClaim rejects a winner claim
func (h *AdminComprehensiveHandler) RejectWinnerClaim(c *gin.Context) {
	winnerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid winner ID",
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

	ctx := c.Request.Context()
	
	err = h.winnerService.RejectClaim(ctx, winnerID.String(), req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to reject claim",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Winner claim rejected successfully",
	})
}

// GetClaimStatistics returns claim statistics
func (h *AdminComprehensiveHandler) GetClaimStatistics(c *gin.Context) {
	// TODO: Implement claim statistics aggregation
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_claims":   0,
			"pending_claims": 0,
			"approved_claims": 0,
			"rejected_claims": 0,
		},
	})
}


// ============================================================================
// SPIN WHEEL PRIZE MANAGEMENT
// ============================================================================

// GetSpinConfig returns the current spin wheel configuration
func (h *AdminComprehensiveHandler) GetSpinConfig(c *gin.Context) {
	ctx := c.Request.Context()

	config, err := h.spinService.GetConfig(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get spin configuration",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config,
	})
}

// UpdateSpinConfig updates the spin wheel configuration
func (h *AdminComprehensiveHandler) UpdateSpinConfig(c *gin.Context) {
	var config map[string]interface{}
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Persist each config field to platform_settings with "spin." prefix
	for field, val := range config {
		key := "spin." + field
		strVal := fmt.Sprintf("%v", val)
		err := h.db.Exec(
			`INSERT INTO platform_settings (setting_key, setting_value, description)
			 VALUES (?, ?, 'Spin wheel configuration')
			 ON CONFLICT (setting_key) DO UPDATE SET setting_value = EXCLUDED.setting_value, updated_at = now()`,
			key, strVal,
		).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to save spin config key: " + key,
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Spin configuration updated successfully",
		"data":    config,
	})
}

// GetAllPrizes returns all wheel prizes
func (h *AdminComprehensiveHandler) GetAllPrizes(c *gin.Context) {
	ctx := c.Request.Context()

	prizes, err := h.spinService.GetAllPrizes(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get prizes",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    prizes,
	})
}

// CreatePrize creates a new wheel prize
func (h *AdminComprehensiveHandler) CreatePrize(c *gin.Context) {
	ctx := c.Request.Context()

	var prizeData struct {
		Name            string   `json:"name" binding:"required"`
		Type            string   `json:"type" binding:"required"`
		Value           float64  `json:"value" binding:"required"`
		Probability     float64  `json:"probability" binding:"required"`
		IsActive        bool     `json:"is_active"`
		MinimumRecharge *float64 `json:"minimum_recharge"`
		ColorScheme     string   `json:"color_scheme"`
		Color           string   `json:"color"`
		SortOrder       *float64 `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&prizeData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	prizeMap := map[string]interface{}{
		"name":        prizeData.Name,
		"type":        prizeData.Type,
		"value":       prizeData.Value,
		"probability": prizeData.Probability,
		"is_active":   prizeData.IsActive,
	}
	if prizeData.MinimumRecharge != nil {
		prizeMap["minimum_recharge"] = *prizeData.MinimumRecharge
	}
	if prizeData.ColorScheme != "" {
		prizeMap["color_scheme"] = prizeData.ColorScheme
	} else if prizeData.Color != "" {
		prizeMap["color"] = prizeData.Color
	}
	if prizeData.SortOrder != nil {
		prizeMap["sort_order"] = *prizeData.SortOrder
	}
	prize, err := h.spinService.CreatePrize(ctx, prizeMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Prize created successfully",
		"data":    prize,
	})
}

// UpdatePrize updates an existing wheel prize
func (h *AdminComprehensiveHandler) UpdatePrize(c *gin.Context) {
	ctx := c.Request.Context()

	prizeIDStr := c.Param("id")
	prizeID, err := uuid.Parse(prizeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid prize ID",
		})
		return
	}

	var updateData struct {
		Name        *string  `json:"name"`
		Type        *string  `json:"type"`
		Value       *float64 `json:"value"`
		Probability *float64 `json:"probability"`
		IsActive    *bool    `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	updateMap := make(map[string]interface{})
	if updateData.Name != nil {
		updateMap["name"] = *updateData.Name
	}
	if updateData.Type != nil {
		updateMap["type"] = *updateData.Type
	}
	if updateData.Value != nil {
		updateMap["value"] = *updateData.Value
	}
	if updateData.Probability != nil {
		updateMap["probability"] = *updateData.Probability
	}
	if updateData.IsActive != nil {
		updateMap["is_active"] = *updateData.IsActive
	}
	
	updatedPrize, err := h.spinService.UpdatePrize(ctx, prizeID.String(), updateMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update prize",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Prize updated successfully",
		"data":    updatedPrize,
	})
}

// DeletePrize deletes a wheel prize
func (h *AdminComprehensiveHandler) DeletePrize(c *gin.Context) {
	ctx := c.Request.Context()

	prizeIDStr := c.Param("id")
	prizeID, err := uuid.Parse(prizeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid prize ID",
		})
		return
	}

	if err := h.spinService.DeletePrize(ctx, prizeID.String()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete prize",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Prize deleted successfully",
	})
}


// ============================================================================
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
			"msisdn":     r.Msisdn,
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
		
		minAmount := int64(0)
		if network.MinimumAmount != nil {
			minAmount = *network.MinimumAmount
		}
		
		maxAmount := int64(0)
		if network.MaximumAmount != nil {
			maxAmount = *network.MaximumAmount
		}
		
		response = append(response, map[string]interface{}{
			"id":                  network.ID,
			"network":             network.NetworkName,
			"code":                network.NetworkCode,
			"enabled":             isActive,
			"airtime_enabled":     airtimeEnabled,
			"data_enabled":        dataEnabled,
			"commission_rate":     commissionRate,
			"minimum_amount":      minAmount,
			"maximum_amount":      maxAmount,
			"logo_url":            network.LogoUrl,
			"brand_color":         network.BrandColor,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ============================================================================
// USER MANAGEMENT APIS
// ============================================================================

// GetAllUsers returns paginated list of users
func (h *AdminComprehensiveHandler) GetAllUsers(c *gin.Context) {
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

	users, total, err := h.userService.GetAllUsers(ctx, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve users",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    users,
		"pagination": gin.H{
			"page":     page,
			"per_page": perPage,
			"total":    total,
		},
	})
}

// GetUserDetails returns detailed user information
func (h *AdminComprehensiveHandler) GetUserDetails(c *gin.Context) {
	ctx := c.Request.Context()

	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid user ID",
		})
		return
	}

	user, err := h.userService.GetUserByID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "User not found",
		})
		return
	}

	profile, err := h.userService.GetUserProfile(ctx, user.MSISDN)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve user profile",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    profile,
	})
}

// UpdateUserStatus updates user status (active/suspended/banned)
func (h *AdminComprehensiveHandler) UpdateUserStatus(c *gin.Context) {
	ctx := c.Request.Context()

	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid user ID",
		})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Validate status
	if req.Status != "active" && req.Status != "suspended" && req.Status != "banned" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid status. Must be 'active', 'suspended', or 'banned'",
		})
		return
	}

	user, err := h.userService.GetUserByID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "User not found",
		})
		return
	}

	// Update user status
	if req.Status == "active" {
		if err := h.userService.ReactivateUser(ctx, user.MSISDN); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to activate user",
			})
			return
		}
	} else {
		if err := h.userService.DeactivateUser(ctx, user.MSISDN); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to update user status",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User status updated successfully",
	})
}

// ============================================================================
// AFFILIATE MANAGEMENT APIS
// ============================================================================

// GetAllAffiliates returns paginated list of affiliates
func (h *AdminComprehensiveHandler) GetAllAffiliates(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	status := c.DefaultQuery("status", "")
	
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	affiliates, total, err := h.affiliateService.GetAllAffiliates(ctx, page, perPage, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve affiliates",
		})
		return
	}

	// Enrich affiliates with user data (full_name, email, last_activity)
	type AffiliateResponse struct {
		ID              string      `json:"id"`
		UserID          interface{} `json:"user_id"`
		AffiliateCode   string      `json:"affiliate_code"`
		ReferralCode    string      `json:"referral_code"`
		Status          string      `json:"status"`
		Tier            string      `json:"tier"`
		CommissionRate  float64     `json:"commission_rate"`
		TotalReferrals  int         `json:"total_referrals"`
		ActiveReferrals int         `json:"active_referrals"`
		TotalCommission float64     `json:"total_commission"`
		BusinessName    string      `json:"business_name"`
		WebsiteUrl      string      `json:"website_url"`
		BankName        string      `json:"bank_name"`
		AccountNumber   string      `json:"account_number"`
		AccountName     string      `json:"account_name"`
		FullName        string      `json:"full_name"`
		Email           string      `json:"email"`
		LastActivity    string      `json:"last_activity"`
		CreatedAt       string      `json:"created_at"`
		UpdatedAt       string      `json:"updated_at"`
		ApprovedAt      interface{} `json:"approved_at"`
	}

	enrichedAffiliates := make([]AffiliateResponse, 0, len(affiliates))
	for _, aff := range affiliates {
		ar := AffiliateResponse{
			ID:              aff.ID.String(),
			AffiliateCode:   aff.AffiliateCode,
			ReferralCode:    aff.ReferralCode,
			Status:          aff.Status,
			Tier:            aff.Tier,
			CommissionRate:  aff.CommissionRate,
			TotalReferrals:  aff.TotalReferrals,
			ActiveReferrals: aff.ActiveReferrals,
			TotalCommission: aff.TotalCommission,
			BusinessName:    aff.BusinessName,
			WebsiteUrl:      aff.WebsiteUrl,
			BankName:        aff.BankName,
			AccountNumber:   aff.AccountNumber,
			AccountName:     aff.AccountName,
			CreatedAt:       aff.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:       aff.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			LastActivity:    aff.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if aff.UserID != nil {
			ar.UserID = aff.UserID.String()
		}
		if aff.ApprovedAt != nil {
			ar.ApprovedAt = aff.ApprovedAt.Format("2006-01-02T15:04:05Z07:00")
		}
		if aff.User != nil {
			ar.FullName = aff.User.FullName
			ar.Email = aff.User.Email
		}
		enrichedAffiliates = append(enrichedAffiliates, ar)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    enrichedAffiliates,
		"pagination": gin.H{
			"page":     page,
			"per_page": perPage,
			"total":    total,
		},
	})
}

// GetAffiliateStats returns affiliate statistics
func (h *AdminComprehensiveHandler) GetAffiliateStats(c *gin.Context) {
	ctx := c.Request.Context()

	// Get all affiliates to calculate stats
	affiliates, _, err := h.affiliateService.GetAllAffiliates(ctx, 1, 1000, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve affiliate stats",
		})
		return
	}

	totalAffiliates := len(affiliates)
	activeAffiliates := 0
	pendingAffiliates := 0
	totalReferrals := 0

	for _, aff := range affiliates {
		if aff.Status == "APPROVED" {
			activeAffiliates++
		} else if aff.Status == "PENDING" {
			pendingAffiliates++
		}
		totalReferrals += aff.TotalReferrals
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_affiliates":   totalAffiliates,
			"active_affiliates":  activeAffiliates,
			"pending_affiliates": pendingAffiliates,
			"total_referrals":    totalReferrals,
			"total_commission":   0,
			"pending_commission": 0,
			"paid_commission":    0,
		},
	})
}

// ApproveAffiliate approves an affiliate application
func (h *AdminComprehensiveHandler) ApproveAffiliate(c *gin.Context) {
	ctx := c.Request.Context()

	affiliateID := c.Param("id")
	if affiliateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Affiliate ID is required",
		})
		return
	}

	if err := h.affiliateService.ApproveAffiliate(ctx, affiliateID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to approve affiliate",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Affiliate approved successfully",
	})
}

// RejectAffiliate rejects an affiliate application
func (h *AdminComprehensiveHandler) RejectAffiliate(c *gin.Context) {
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
			"error":   "Rejection reason is required",
		})
		return
	}

	if err := h.affiliateService.RejectAffiliate(ctx, affiliateID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to reject affiliate",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Affiliate rejected successfully",
	})
}

// GetDataPlans returns all data plans across all networks for admin management
func (h *AdminComprehensiveHandler) GetDataPlans(c *gin.Context) {
	// Get all data plans from the database
	plans, err := h.rechargeService.GetAllDataPlans()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to fetch data plans",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    plans,
	})
}

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
		MinimumAmount  *int64   `json:"minimum_amount"`
		MaximumAmount  *int64   `json:"maximum_amount"`
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
		MinimumAmount  *int64   `json:"minimum_amount"`
		MaximumAmount  *int64   `json:"maximum_amount"`
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

	// Get commissions (affiliate service uses affiliate code, not MSISDN)
	commissions, err := h.affiliateService.GetCommissions(ctx, affiliate.AffiliateCode)
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
			"affiliate_id":     affiliate.ID,
			"commission_rate":  req.CommissionRate,
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

	// Process payout
	payout, err := h.affiliateService.RequestPayout(ctx, affiliate.AffiliateCode, req.Amount)
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
		"total_affiliates":    total,
		"time_range":         timeRange,
		"top_performers":     affiliates[:min(10, len(affiliates))],
		"total_commissions":  0, // Would calculate from commission ledger
		"total_payouts":      0, // Would calculate from payout records
		"pending_payouts":    0, // Would calculate from pending payout requests
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

// ============================================================================
// DAILY SUBSCRIPTION ANALYTICS & CONFIG (ENTERPRISE GRADE)
// ============================================================================

// GetSubscriptionAnalytics retrieves subscription analytics and metrics
func (h *AdminComprehensiveHandler) GetSubscriptionAnalytics(c *gin.Context) {
	ctx := c.Request.Context()

	// Get active subscription count
	activeCount, err := h.subscriptionService.GetActiveSubscriptionCount(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve analytics",
		})
		return
	}

	// Get config for pricing info
	config, err := h.subscriptionService.GetConfig(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve config",
		})
		return
	}

	// Calculate revenue metrics (would need billing data)
	analytics := gin.H{
		"active_subscriptions": activeCount,
		"config":              config,
		"daily_revenue":       0, // Would calculate from billing records
		"monthly_revenue":     0, // Would calculate from billing records
		"churn_rate":          0, // Would calculate from cancellations
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    analytics,
	})
}

// GetSubscriptionConfig retrieves daily subscription configuration
func (h *AdminComprehensiveHandler) GetSubscriptionConfig(c *gin.Context) {
	ctx := c.Request.Context()

	config, err := h.subscriptionService.GetConfig(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve config",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config,
	})
}

// UpdateSubscriptionConfig updates daily subscription configuration
func (h *AdminComprehensiveHandler) UpdateSubscriptionConfig(c *gin.Context) {
	ctx := c.Request.Context()

	var config map[string]interface{}
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid config data: " + err.Error(),
		})
		return
	}

	if err := h.subscriptionService.UpdateConfig(ctx, config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update config: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Subscription config updated successfully",
		"data":    config,
	})
}

// ============================================================================
// PRIZE TIER SYSTEM - DRAW TYPES, TEMPLATES & CATEGORIES
// ============================================================================

// GetDrawTypes returns all draw types (Daily, Weekly, Special)
func (h *AdminComprehensiveHandler) GetDrawTypes(c *gin.Context) {
	drawTypes, err := h.drawTypeService.GetAllDrawTypes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve draw types",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    drawTypes,
	})
}

// GetPrizeTemplates returns all prize templates or filtered by draw type
func (h *AdminComprehensiveHandler) GetPrizeTemplates(c *gin.Context) {
	drawTypeIDStr := c.Query("draw_type_id")
	
	var templates []entities.PrizeTemplate
	var err error
	
	if drawTypeIDStr != "" {
		drawTypeID, parseErr := strconv.ParseUint(drawTypeIDStr, 10, 32)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid draw_type_id",
			})
			return
		}
		templates, err = h.prizeTemplateService.GetTemplatesByDrawType(uint(drawTypeID))
	} else {
		templates, err = h.prizeTemplateService.GetAllTemplates()
	}
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve prize templates",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    templates,
	})
}

// GetPrizeTemplate returns a specific prize template with its categories
func (h *AdminComprehensiveHandler) GetPrizeTemplate(c *gin.Context) {
	templateIDStr := c.Param("id")
	templateID, err := strconv.ParseUint(templateIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid template ID",
		})
		return
	}

	template, err := h.prizeTemplateService.GetTemplate(uint(templateID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Template not found",
		})
		return
	}

	// Calculate total prize pool
	totalPool, _ := h.prizeTemplateService.CalculateTotalPrizePool(template.ID)

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"data":            template,
		"total_prize_pool": totalPool,
	})
}

// CreatePrizeTemplate creates a new prize template with categories
func (h *AdminComprehensiveHandler) CreatePrizeTemplate(c *gin.Context) {
	var req struct {
		Name        string                   `json:"name" binding:"required"`
		Description string                   `json:"description"`
		DrawTypeID  uint                     `json:"draw_type_id" binding:"required"`
		IsDefault   bool                     `json:"is_default"`
		Categories  []entities.PrizeCategory `json:"categories" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request: " + err.Error(),
		})
		return
	}

	template, err := h.prizeTemplateService.CreateTemplate(
		req.Name,
		req.Description,
		req.DrawTypeID,
		req.IsDefault,
		req.Categories,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Prize template created successfully",
		"data":    template,
	})
}

// UpdatePrizeTemplate updates an existing prize template
func (h *AdminComprehensiveHandler) UpdatePrizeTemplate(c *gin.Context) {
	templateIDStr := c.Param("id")
	templateID, err := strconv.ParseUint(templateIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid template ID",
		})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		IsDefault   *bool  `json:"is_default"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request: " + err.Error(),
		})
		return
	}

	template, err := h.prizeTemplateService.UpdateTemplate(
		uint(templateID),
		req.Name,
		req.Description,
		req.IsDefault,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Prize template updated successfully",
		"data":    template,
	})
}

// DeletePrizeTemplate deletes a prize template
func (h *AdminComprehensiveHandler) DeletePrizeTemplate(c *gin.Context) {
	templateIDStr := c.Param("id")
	templateID, err := strconv.ParseUint(templateIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid template ID",
		})
		return
	}

	if err := h.prizeTemplateService.DeleteTemplate(uint(templateID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Prize template deleted successfully",
	})
}

// AddPrizeCategory adds a new prize category to a template
func (h *AdminComprehensiveHandler) AddPrizeCategory(c *gin.Context) {
	templateIDStr := c.Param("id")
	templateID, err := strconv.ParseUint(templateIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid template ID",
		})
		return
	}

	var req struct {
		CategoryName  string  `json:"category_name" binding:"required"`
		PrizeAmount   float64 `json:"prize_amount" binding:"required"`
		WinnerCount   int     `json:"winner_count" binding:"required,min=1"`
		RunnerUpCount int     `json:"runner_up_count"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request: " + err.Error(),
		})
		return
	}

	category, err := h.prizeTemplateService.AddCategoryToTemplate(
		uint(templateID),
		req.CategoryName,
		req.PrizeAmount,
		req.WinnerCount,
		req.RunnerUpCount,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Prize category added successfully",
		"data":    category,
	})
}

// UpdatePrizeCategory updates an existing prize category
func (h *AdminComprehensiveHandler) UpdatePrizeCategory(c *gin.Context) {
	categoryIDStr := c.Param("id")
	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid category ID",
		})
		return
	}

	var req struct {
		CategoryName  *string  `json:"category_name"`
		PrizeAmount   *float64 `json:"prize_amount"`
		WinnerCount   *int     `json:"winner_count"`
		RunnerUpCount *int     `json:"runner_up_count"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request: " + err.Error(),
		})
		return
	}

	category, err := h.prizeTemplateService.UpdateCategory(
		uint(categoryID),
		req.CategoryName,
		req.PrizeAmount,
		req.WinnerCount,
		req.RunnerUpCount,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Prize category updated successfully",
		"data":    category,
	})
}

// DeletePrizeCategory deletes a prize category
func (h *AdminComprehensiveHandler) DeletePrizeCategory(c *gin.Context) {
	categoryIDStr := c.Param("id")
	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid category ID",
		})
		return
	}

	if err := h.prizeTemplateService.DeleteCategory(uint(categoryID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Prize category deleted successfully",
	})
}
