package handlers

// admin_missing_routes_handler.go
// Implements the admin endpoints that the frontend calls but were not yet
// wired to backend routes.  All handlers are methods on AdminComprehensiveHandler.

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"rechargemax/internal/application/services"
)

// ──────────────────────────────────────────────────────────────────────────────
// USER MANAGEMENT — bare /:id, DELETE, suspend/activate, points-history
// ──────────────────────────────────────────────────────────────────────────────

// GetUser returns a single user by ID (alias to GetUserDetails without /details suffix)
func (h *AdminComprehensiveHandler) GetUser(c *gin.Context) {
	h.GetUserDetails(c)
}

// UpdateUser allows updating a user's profile fields (PUT /admin/users/:id)
func (h *AdminComprehensiveHandler) UpdateUser(c *gin.Context) {
	h.UpdateUserStatus(c)
}

// DeleteUser soft-deletes / deactivates a user account
func (h *AdminComprehensiveHandler) DeleteUser(c *gin.Context) {
	ctx := c.Request.Context()

	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid user ID"})
		return
	}

	user, err := h.userService.GetUserByID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "User not found"})
		return
	}

	if err := h.userService.DeactivateUser(ctx, user.MSISDN); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to deactivate user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "User deactivated successfully"})
}

// SuspendUser suspends a user account (POST /admin/users/:id/suspend)
func (h *AdminComprehensiveHandler) SuspendUser(c *gin.Context) {
	ctx := c.Request.Context()

	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid user ID"})
		return
	}

	user, err := h.userService.GetUserByID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "User not found"})
		return
	}

	if err := h.userService.DeactivateUser(ctx, user.MSISDN); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to suspend user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "User suspended successfully"})
}

// ActivateUser reactivates a suspended user (POST /admin/users/:id/activate)
func (h *AdminComprehensiveHandler) ActivateUser(c *gin.Context) {
	ctx := c.Request.Context()

	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid user ID"})
		return
	}

	user, err := h.userService.GetUserByID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "User not found"})
		return
	}

	if err := h.userService.ReactivateUser(ctx, user.MSISDN); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to activate user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "User activated successfully"})
}

// GetUserPointsHistory returns a user's points transaction history
// GET /admin/users/:id/points-history
func (h *AdminComprehensiveHandler) GetUserPointsHistory(c *gin.Context) {
	ctx := c.Request.Context()

	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid user ID"})
		return
	}

	history, err := h.pointsService.GetPointsHistory(ctx, &userID, "", nil, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to retrieve points history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    history,
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// WINNER MANAGEMENT — list, detail, payout, mark-shipped, notification, runner-up
// ──────────────────────────────────────────────────────────────────────────────

// GetAllWinners returns paginated list of all winners
// GET /admin/winners
func (h *AdminComprehensiveHandler) GetAllWinners(c *gin.Context) {
	ctx := c.Request.Context()

	page := 1
	perPage := 20
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if l := c.Query("per_page"); l != "" {
		fmt.Sscanf(l, "%d", &perPage)
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	drawID := c.DefaultQuery("draw_id", "")

	winners, total, err := h.winnerService.GetAllWinners(ctx, page, perPage, drawID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to retrieve winners"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    winners,
		"pagination": gin.H{
			"page":     page,
			"per_page": perPage,
			"total":    total,
		},
	})
}

// GetWinnerByID returns a single winner record
// GET /admin/winners/:id
func (h *AdminComprehensiveHandler) GetWinnerByID(c *gin.Context) {
	ctx := c.Request.Context()

	winnerIDStr := c.Param("id")
	if _, err := uuid.Parse(winnerIDStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid winner ID"})
		return
	}

	winner, err := h.winnerService.GetWinnerByID(ctx, winnerIDStr, "")
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Winner not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": winner})
}

// ProcessWinnerPayout processes a cash payout for a winner
// POST /admin/winners/:id/process-payout
func (h *AdminComprehensiveHandler) ProcessWinnerPayout(c *gin.Context) {
	ctx := c.Request.Context()

	winnerIDStr := c.Param("id")
	winnerID, err := uuid.Parse(winnerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid winner ID"})
		return
	}

	var req struct {
		BankName      string `json:"bank_name" binding:"required"`
		AccountNumber string `json:"account_number" binding:"required"`
		AccountName   string `json:"account_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if err := h.winnerService.ProcessCashPayout(ctx, winnerID, req.BankName, req.AccountNumber, req.AccountName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to process payout: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Payout processed successfully"})
}

// MarkWinnerShipped marks a physical prize as shipped
// POST /admin/winners/:id/mark-shipped
func (h *AdminComprehensiveHandler) MarkWinnerShipped(c *gin.Context) {
	ctx := c.Request.Context()

	winnerIDStr := c.Param("id")
	winnerID, err := uuid.Parse(winnerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid winner ID"})
		return
	}

	var req struct {
		ShippingAddress string `json:"shipping_address" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if err := h.winnerService.ProcessGoodsShipment(ctx, winnerID, req.ShippingAddress); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to mark as shipped: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Prize marked as shipped"})
}

// SendWinnerNotification sends an SMS notification to a winner
// POST /admin/winners/:id/send-notification
func (h *AdminComprehensiveHandler) SendWinnerNotification(c *gin.Context) {
	ctx := c.Request.Context()

	winnerIDStr := c.Param("id")
	if _, err := uuid.Parse(winnerIDStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid winner ID"})
		return
	}

	winner, err := h.winnerService.GetWinnerByID(ctx, winnerIDStr, "")
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Winner not found"})
		return
	}

	var req struct {
		Message string `json:"message"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.Message == "" {
		req.Message = "Congratulations! You have won a prize on RechargeMax. Please log in to claim your prize."
	}

	// Look up the user_id from the users table using the winner's MSISDN
	var userID *string
	var row struct {
		ID string `gorm:"column:id"`
	}
	if err := h.db.WithContext(ctx).Raw(
		"SELECT id FROM users WHERE msisdn = ? LIMIT 1", winner.MSISDN,
	).Scan(&row).Error; err == nil && row.ID != "" {
		userID = &row.ID
	}

	// Insert into user_notifications (correct table + schema)
	if err := h.db.WithContext(ctx).Exec(
		`INSERT INTO user_notifications
		   (id, user_id, title, body, notification_type, is_read, created_at, updated_at)
		 VALUES
		   (gen_random_uuid(), ?, 'Prize Notification', ?, 'prize', false, NOW(), NOW())`,
		userID, req.Message,
	).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to queue notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Notification queued successfully"})
}

// InvokeWinnerRunnerUp invokes the runner-up for a winner who failed to claim
// POST /admin/winners/:id/invoke-runner-up
func (h *AdminComprehensiveHandler) InvokeWinnerRunnerUp(c *gin.Context) {
	ctx := c.Request.Context()

	winnerIDStr := c.Param("id")
	winnerID, err := uuid.Parse(winnerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid winner ID"})
		return
	}

	if err := h.winnerService.RetryProvisioning(ctx, winnerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to invoke runner-up: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Runner-up invoked successfully"})
}

// ──────────────────────────────────────────────────────────────────────────────
// RECHARGE MANAGEMENT — detail, refund, mark-success, mark-failed
// ──────────────────────────────────────────────────────────────────────────────

// GetRechargeByID returns a single recharge transaction
// GET /admin/recharge/transactions/:id
func (h *AdminComprehensiveHandler) GetRechargeByID(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid recharge ID"})
		return
	}

	recharge, err := h.rechargeService.GetRechargeByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Recharge transaction not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": recharge})
}

// RefundRecharge processes a refund for a recharge transaction
// POST /admin/recharge/:id/refund
func (h *AdminComprehensiveHandler) RefundRecharge(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid recharge ID"})
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Refund reason is required"})
		return
	}

	// Load the recharge record and update its status to CANCELLED (refunded).
	// DB CHECK constraint only allows: PENDING, PROCESSING, SUCCESS, FAILED, CANCELLED
	recharge, err := h.rechargeService.GetRechargeByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Recharge not found"})
		return
	}
	recharge.Status = "CANCELLED"
	recharge.FailureReason = req.Reason
	recharge.UpdatedAt = time.Now()

	if err := h.rechargeService.UpdateRecharge(ctx, recharge); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to process refund: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Refund processed successfully"})
}

// MarkRechargeSuccess manually marks a recharge as successful
// POST /admin/recharge/:id/mark-success
func (h *AdminComprehensiveHandler) MarkRechargeSuccess(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid recharge ID"})
		return
	}

	recharge, err := h.rechargeService.GetRechargeByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Recharge not found"})
		return
	}
	recharge.Status = "SUCCESS"
	recharge.UpdatedAt = time.Now()

	if err := h.rechargeService.UpdateRecharge(ctx, recharge); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to mark as success: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Recharge marked as successful"})
}

// MarkRechargeFailed manually marks a recharge as failed
// POST /admin/recharge/:id/mark-failed
func (h *AdminComprehensiveHandler) MarkRechargeFailed(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid recharge ID"})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.Reason == "" {
		req.Reason = "Manually marked as failed by admin"
	}

	recharge, err := h.rechargeService.GetRechargeByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Recharge not found"})
		return
	}
	recharge.Status = "FAILED"
	recharge.FailureReason = req.Reason
	recharge.UpdatedAt = time.Now()

	if err := h.rechargeService.UpdateRecharge(ctx, recharge); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to mark as failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Recharge marked as failed"})
}

// ──────────────────────────────────────────────────────────────────────────────
// SUBSCRIPTION MANAGEMENT — pause, resume, per-subscription billings
// ──────────────────────────────────────────────────────────────────────────────

// PauseDailySubscription pauses an active daily subscription
// POST /admin/daily-subscriptions/:id/pause
func (h *AdminComprehensiveHandler) PauseDailySubscription(c *gin.Context) {
	ctx := c.Request.Context()

	subIDStr := c.Param("id")
	subID, err := uuid.Parse(subIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid subscription ID"})
		return
	}

	if err := h.subscriptionTierService.PauseSubscription(ctx, subID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to pause subscription: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Subscription paused successfully"})
}

// ResumeDailySubscription resumes a paused daily subscription
// POST /admin/daily-subscriptions/:id/resume
func (h *AdminComprehensiveHandler) ResumeDailySubscription(c *gin.Context) {
	ctx := c.Request.Context()

	subIDStr := c.Param("id")
	subID, err := uuid.Parse(subIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid subscription ID"})
		return
	}

	if err := h.subscriptionTierService.ResumeSubscription(ctx, subID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to resume subscription: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Subscription resumed successfully"})
}

// GetSubscriptionBillingsByID returns billing records for a specific subscription
// GET /admin/daily-subscriptions/:id/billings
func (h *AdminComprehensiveHandler) GetSubscriptionBillingsByID(c *gin.Context) {
	ctx := c.Request.Context()

	subIDStr := c.Param("id")
	if _, err := uuid.Parse(subIDStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid subscription ID"})
		return
	}

	page := 1
	limit := 20
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	type billing struct {
		ID         string    `json:"id"          gorm:"column:id"`
		MSISDN     string    `json:"msisdn"      gorm:"column:msisdn"`
		Amount     int64     `json:"amount"      gorm:"column:amount"`
		Status     string    `json:"status"      gorm:"column:status"`
		CreatedAt  time.Time `json:"created_at"  gorm:"column:created_at"`
	}
	var billings []billing
	var total int64

	q := h.db.WithContext(ctx).Table("transactions").
		Where("subscription_id = ?", subIDStr)
	q.Count(&total)
	q.Order("created_at DESC").Offset(offset).Limit(limit).Scan(&billings)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    billings,
		"meta":    gin.H{"page": page, "limit": limit, "total": total},
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// SUBSCRIPTION TIER — PATCH toggle-active
// ──────────────────────────────────────────────────────────────────────────────

// ToggleSubscriptionTier toggles the active state of a subscription tier
// PATCH /admin/subscription-tiers/:id/toggle-active
func (h *AdminComprehensiveHandler) ToggleSubscriptionTier(c *gin.Context) {
	ctx := c.Request.Context()

	tierIDStr := c.Param("id")
	tierID, err := uuid.Parse(tierIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid tier ID"})
		return
	}

	// Get all tiers to find the current is_active state
	tiers, err := h.subscriptionTierService.GetAllTiers(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to retrieve tier"})
		return
	}

	var currentActive bool
	found := false
	for _, t := range tiers {
		if t.ID == tierID {
			currentActive = t.IsActive
			found = true
			break
		}
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Subscription tier not found"})
		return
	}

	newActive := !currentActive
	// Re-fetch the full tier to preserve name/description/entries/sortOrder
	allTiers, err2 := h.subscriptionTierService.GetAllTiers(ctx)
	if err2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to re-fetch tier"})
		return
	}
	var targetTier *struct{ Name, Description string; Entries, SortOrder int }
	for _, t := range allTiers {
		if t.ID == tierID {
			targetTier = &struct{ Name, Description string; Entries, SortOrder int }{t.Name, t.Description, t.Entries, t.SortOrder}
			break
		}
	}
	if targetTier == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Tier disappeared"})
		return
	}
	if _, err := h.subscriptionTierService.UpdateTier(ctx, tierID, targetTier.Name, targetTier.Description, targetTier.Entries, targetTier.SortOrder, newActive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to toggle tier status"})
		return
	}

	msg := "Subscription tier deactivated"
	if newActive {
		msg = "Subscription tier activated"
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": msg,
		"data":    gin.H{"is_active": newActive},
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// AFFILIATE MANAGEMENT — CRUD + payout-history
// ──────────────────────────────────────────────────────────────────────────────

// CreateAffiliate creates a new affiliate (admin-initiated)
// POST /admin/affiliates
func (h *AdminComprehensiveHandler) CreateAffiliate(c *gin.Context) {
	ctx := c.Request.Context()

	var req services.RegisterAffiliateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	affiliate, err := h.affiliateService.RegisterAffiliate(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create affiliate: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": affiliate})
}

// UpdateAffiliate updates affiliate commission rate / status
// PUT /admin/affiliates/:id
func (h *AdminComprehensiveHandler) UpdateAffiliate(c *gin.Context) {
	// Delegate to the existing UpdateAffiliateCommissionRate handler
	h.UpdateAffiliateCommissionRate(c)
}

// DeleteAffiliate removes / suspends an affiliate
// DELETE /admin/affiliates/:id
func (h *AdminComprehensiveHandler) DeleteAffiliate(c *gin.Context) {
	ctx := c.Request.Context()

	affiliateID := c.Param("id")
	if affiliateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Affiliate ID is required"})
		return
	}

	if err := h.affiliateService.SuspendAffiliate(ctx, affiliateID, "Deleted by admin"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to remove affiliate"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Affiliate removed successfully"})
}

// GetAffiliatePayoutHistory returns payout history for a specific affiliate
// GET /admin/affiliates/:id/payout-history
func (h *AdminComprehensiveHandler) GetAffiliatePayoutHistory(c *gin.Context) {
	// Delegate to existing GetAffiliatePayouts handler (same data, different path)
	h.GetAffiliatePayouts(c)
}
