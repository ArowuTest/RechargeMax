package handlers

import (
	"net/http"
	
	"github.com/gin-gonic/gin"
	
	"rechargemax/internal/application/services"
)

type AdminHandler struct {
	drawService         *services.DrawService
	winnerService       *services.WinnerService
	userService         *services.UserService
	affiliateService    *services.AffiliateService
	subscriptionService *services.SubscriptionService
	spinService         *services.SpinService
	networkService      *services.NetworkConfigService
	rechargeService     *services.RechargeService
}

func NewAdminHandler(
	drawService *services.DrawService,
	winnerService *services.WinnerService,
	userService *services.UserService,
) *AdminHandler {
	return &AdminHandler{
		drawService:   drawService,
		winnerService: winnerService,
		userService:   userService,
	}
}

// SetServices allows setting additional services after initialization
func (h *AdminHandler) SetServices(
	affiliateService *services.AffiliateService,
	subscriptionService *services.SubscriptionService,
	spinService *services.SpinService,
	networkService *services.NetworkConfigService,
	rechargeService *services.RechargeService,
) {
	h.affiliateService = affiliateService
	h.subscriptionService = subscriptionService
	h.spinService = spinService
	h.networkService = networkService
	h.rechargeService = rechargeService
}

// GetDashboardStats returns admin dashboard statistics
func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
	ctx := c.Request.Context()
	
	// Get total users count
	totalUsers, _ := h.userService.GetUserCount(ctx)
	
	// Get users list to count active ones
	_, totalUsersFromList, _ := h.userService.GetAllUsers(ctx, 1, 1)
	if totalUsers == 0 {
		totalUsers = totalUsersFromList
	}
	
	// Get affiliate stats
	totalAffiliates := int64(0)
	pendingAffiliates := int64(0)
	totalCommissions := float64(0)
	if h.affiliateService != nil {
		affiliates, _, _ := h.affiliateService.GetAllAffiliates(ctx, 1, 1000, "")
		if affiliates != nil {
			totalAffiliates = int64(len(affiliates))
			for _, a := range affiliates {
				if a.Status == "PENDING" {
					pendingAffiliates++
				}
				totalCommissions += float64(a.TotalCommission)
			}
		}
	}
	
	// Get active draws count
	activeDraws := int64(0)
	if h.drawService != nil {
		draws, _ := h.drawService.GetActiveDraws(ctx)
		if draws != nil {
			activeDraws = int64(len(draws))
		}
	}

	// Get pending winner claims count
	pendingClaims := int64(0)
	if h.winnerService != nil {
		_, totalWinners, _ := h.winnerService.GetAllWinners(ctx, 1, 1000, "")
		pendingClaims = totalWinners
	}

	stats := gin.H{
		"total_users":          totalUsers,
		"active_draws":         activeDraws,
		"pending_claims":       pendingClaims,
		"active_subscriptions": 0,
		"total_recharges":      0,
		"total_revenue":        0,
		"total_affiliates":     totalAffiliates,
		"pending_affiliates":   pendingAffiliates,
		"total_commissions":    totalCommissions,
		"new_users_today":      0,
		"transactions_today":   0,
		"today_revenue":        0,
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetUsers returns list of users for admin
func (h *AdminHandler) GetUsers(c *gin.Context) {
	ctx := c.Request.Context()
	
	users, _, err := h.userService.GetAllUsers(ctx, 1, 1000)
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
	})
}

// GetDraws returns list of draws for admin
func (h *AdminHandler) GetDraws(c *gin.Context) {
	ctx := c.Request.Context()
	
	// Get all draws with pagination (page 1, limit 100)
	draws, total, err := h.drawService.GetDraws(ctx, 1, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve draws",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    draws,
		"total":   total,
	})
}
