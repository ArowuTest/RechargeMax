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
	// TODO: Implement dashboard stats aggregation
	stats := gin.H{
		"total_users":         0,
		"active_subscriptions": 0,
		"total_recharges":     0,
		"total_revenue":       0,
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
