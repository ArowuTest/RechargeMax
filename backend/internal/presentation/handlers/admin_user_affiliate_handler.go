package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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
