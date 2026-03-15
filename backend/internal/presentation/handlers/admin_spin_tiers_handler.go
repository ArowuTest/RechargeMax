package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"rechargemax/internal/application/services"
)

// AdminSpinTiersHandler handles admin spin-tier CRUD.
type AdminSpinTiersHandler struct {
	spinTiersSvc *services.SpinTiersService
}

// NewAdminSpinTiersHandler creates a new AdminSpinTiersHandler.
func NewAdminSpinTiersHandler(spinTiersSvc *services.SpinTiersService) *AdminSpinTiersHandler {
	return &AdminSpinTiersHandler{spinTiersSvc: spinTiersSvc}
}

// GetAllTiers returns all spin tiers (active + inactive), ordered by sort_order.
func (h *AdminSpinTiersHandler) GetAllTiers(c *gin.Context) {
	tiers, err := h.spinTiersSvc.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch spin tiers", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "tiers": tiers, "count": len(tiers)})
}

// GetTierByID returns a single spin tier.
func (h *AdminSpinTiersHandler) GetTierByID(c *gin.Context) {
	tier, err := h.spinTiersSvc.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tier not found", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "tier": tier})
}

// updateTierRequest carries optional-update fields for a tier.
type updateTierRequest struct {
	TierDisplayName *string `json:"tier_display_name"`
	MinDailyAmount  *int64  `json:"min_daily_amount"`
	MaxDailyAmount  *int64  `json:"max_daily_amount"`
	SpinsPerDay     *int    `json:"spins_per_day"`
	TierColor       *string `json:"tier_color"`
	TierIcon        *string `json:"tier_icon"`
	TierBadge       *string `json:"tier_badge"`
	Description     *string `json:"description"`
	SortOrder       *int    `json:"sort_order"`
	IsActive        *bool   `json:"is_active"`
}

// UpdateTier applies partial updates to a spin tier.
func (h *AdminSpinTiersHandler) UpdateTier(c *gin.Context) {
	adminID, _ := c.Get("admin_id")
	aid, _ := adminID.(string)

	var req updateTierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	tier, err := h.spinTiersSvc.Update(c.Request.Context(), c.Param("id"), services.UpdateSpinTierRequest{
		TierDisplayName: req.TierDisplayName,
		MinDailyAmount:  req.MinDailyAmount,
		MaxDailyAmount:  req.MaxDailyAmount,
		SpinsPerDay:     req.SpinsPerDay,
		TierColor:       req.TierColor,
		TierIcon:        req.TierIcon,
		TierBadge:       req.TierBadge,
		Description:     req.Description,
		SortOrder:       req.SortOrder,
		IsActive:        req.IsActive,
	}, aid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Tier updated successfully", "tier": tier})
}

// createTierRequest carries required fields for a new tier.
type createTierRequest struct {
	TierName        string `json:"tier_name" binding:"required"`
	TierDisplayName string `json:"tier_display_name" binding:"required"`
	MinDailyAmount  int64  `json:"min_daily_amount" binding:"required"`
	MaxDailyAmount  int64  `json:"max_daily_amount" binding:"required"`
	SpinsPerDay     int    `json:"spins_per_day" binding:"required"`
	TierColor       string `json:"tier_color"`
	TierIcon        string `json:"tier_icon"`
	TierBadge       string `json:"tier_badge"`
	Description     string `json:"description"`
	SortOrder       int    `json:"sort_order"`
}

// CreateTier inserts a new spin tier.
func (h *AdminSpinTiersHandler) CreateTier(c *gin.Context) {
	adminID, _ := c.Get("admin_id")
	aid, _ := adminID.(string)

	var req createTierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	tier, err := h.spinTiersSvc.Create(c.Request.Context(), services.CreateSpinTierRequest{
		TierName:        req.TierName,
		TierDisplayName: req.TierDisplayName,
		MinDailyAmount:  req.MinDailyAmount,
		MaxDailyAmount:  req.MaxDailyAmount,
		SpinsPerDay:     req.SpinsPerDay,
		TierColor:       req.TierColor,
		TierIcon:        req.TierIcon,
		TierBadge:       req.TierBadge,
		Description:     req.Description,
		SortOrder:       req.SortOrder,
	}, aid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Tier created successfully", "tier": tier})
}

// DeleteTier soft-deletes a spin tier.
func (h *AdminSpinTiersHandler) DeleteTier(c *gin.Context) {
	tier, err := h.spinTiersSvc.Delete(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Tier deleted successfully", "tier": tier})
}

// ValidateTierConfiguration is not routed but kept for potential future use.
// It can be called via a dedicated admin endpoint if needed.
