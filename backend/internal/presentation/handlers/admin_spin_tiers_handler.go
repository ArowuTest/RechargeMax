package handlers

import (
	"net/http"
	"rechargemax/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AdminSpinTiersHandler handles admin operations for spin tiers
type AdminSpinTiersHandler struct {
	db         *gorm.DB
	calculator *utils.SpinTierCalculatorDB
}

// NewAdminSpinTiersHandler creates a new admin spin tiers handler
func NewAdminSpinTiersHandler(db *gorm.DB) *AdminSpinTiersHandler {
	return &AdminSpinTiersHandler{
		db:         db,
		calculator: utils.NewSpinTierCalculatorDB(db),
	}
}

// GetAllTiers returns all spin tiers (including inactive)
// GET /api/admin/spin-tiers
func (h *AdminSpinTiersHandler) GetAllTiers(c *gin.Context) {
	var tiers []utils.SpinTierDB
	
	if err := h.db.Order("sort_order ASC").Find(&tiers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch spin tiers",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"tiers":   tiers,
		"count":   len(tiers),
	})
}

// GetActiveTiers returns only active spin tiers
// GET /api/admin/spin-tiers/active
func (h *AdminSpinTiersHandler) GetActiveTiers(c *gin.Context) {
	tiers, err := h.calculator.GetAllTiersFromDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch active spin tiers",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"tiers":   tiers,
		"count":   len(tiers),
	})
}

// GetTierByID returns a specific tier by ID
// GET /api/admin/spin-tiers/:id
func (h *AdminSpinTiersHandler) GetTierByID(c *gin.Context) {
	tierID := c.Param("id")

	tier, err := h.calculator.GetTierByIDFromDB(tierID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Tier not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"tier":    tier,
	})
}

// UpdateTierRequest represents the request body for updating a tier
type UpdateTierRequest struct {
	TierDisplayName *string `json:"tier_display_name"`
	MinDailyAmount  *int64  `json:"min_daily_amount"`  // Amount in kobo
	MaxDailyAmount  *int64  `json:"max_daily_amount"`  // Amount in kobo
	SpinsPerDay     *int    `json:"spins_per_day"`
	TierColor       *string `json:"tier_color"`
	TierIcon        *string `json:"tier_icon"`
	TierBadge       *string `json:"tier_badge"`
	Description     *string `json:"description"`
	SortOrder       *int    `json:"sort_order"`
	IsActive        *bool   `json:"is_active"`
}

// UpdateTier updates a spin tier
// PUT /api/admin/spin-tiers/:id
func (h *AdminSpinTiersHandler) UpdateTier(c *gin.Context) {
	tierID := c.Param("id")
	
	// Get admin ID from context (set by auth middleware)
	adminID, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Admin ID not found in context",
		})
		return
	}

	var req UpdateTierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Fetch existing tier
	_, err := h.calculator.GetTierByIDFromDB(tierID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Tier not found",
			"details": err.Error(),
		})
		return
	}

	// Update fields if provided
	updates := make(map[string]interface{})
	
	if req.TierDisplayName != nil {
		updates["tier_display_name"] = *req.TierDisplayName
	}
	if req.MinDailyAmount != nil {
		if *req.MinDailyAmount < 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Minimum daily amount cannot be negative",
			})
			return
		}
		updates["min_daily_amount"] = *req.MinDailyAmount
	}
	if req.MaxDailyAmount != nil {
		if *req.MaxDailyAmount <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Maximum daily amount must be positive",
			})
			return
		}
		updates["max_daily_amount"] = *req.MaxDailyAmount
	}
	if req.SpinsPerDay != nil {
		if *req.SpinsPerDay <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Spins per day must be positive",
			})
			return
		}
		updates["spins_per_day"] = *req.SpinsPerDay
	}
	if req.TierColor != nil {
		updates["tier_color"] = *req.TierColor
	}
	if req.TierIcon != nil {
		updates["tier_icon"] = *req.TierIcon
	}
	if req.TierBadge != nil {
		updates["tier_badge"] = *req.TierBadge
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.SortOrder != nil {
		if *req.SortOrder < 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Sort order cannot be negative",
			})
			return
		}
		updates["sort_order"] = *req.SortOrder
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	
	updates["updated_by"] = adminID

	// Perform update
	if err := h.db.Model(&utils.SpinTierDB{}).Where("id = ?", tierID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update tier",
			"details": err.Error(),
		})
		return
	}

	// Validate tier configuration after update
	if err := h.calculator.ValidateTierConfiguration(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tier configuration validation failed",
			"details": err.Error(),
			"message": "The update created overlapping or gapped tier ranges. Please adjust the amounts.",
		})
		return
	}

	// Fetch updated tier
	updatedTier, _ := h.calculator.GetTierByIDFromDB(tierID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tier updated successfully",
		"tier":    updatedTier,
	})
}

// CreateTierRequest represents the request body for creating a new tier
type CreateTierRequest struct {
	TierName        string `json:"tier_name" binding:"required"`
	TierDisplayName string `json:"tier_display_name" binding:"required"`
	MinDailyAmount  int64  `json:"min_daily_amount" binding:"required"`  // Amount in kobo
	MaxDailyAmount  int64  `json:"max_daily_amount" binding:"required"`  // Amount in kobo
	SpinsPerDay     int    `json:"spins_per_day" binding:"required"`
	TierColor       string `json:"tier_color"`
	TierIcon        string `json:"tier_icon"`
	TierBadge       string `json:"tier_badge"`
	Description     string `json:"description"`
	SortOrder       int    `json:"sort_order"`
}

// CreateTier creates a new spin tier
// POST /api/admin/spin-tiers
func (h *AdminSpinTiersHandler) CreateTier(c *gin.Context) {
	// Get admin ID from context
	adminIDVal, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Admin ID not found in context",
		})
		return
	}
	adminID, _ := adminIDVal.(string)

	var req CreateTierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate amounts
	if req.MinDailyAmount < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Minimum daily amount cannot be negative",
		})
		return
	}
	if req.MaxDailyAmount <= req.MinDailyAmount {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Maximum daily amount must be greater than minimum daily amount",
		})
		return
	}
	if req.SpinsPerDay <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Spins per day must be positive",
		})
		return
	}

	// Create tier, setting the admin as creator
	tierID := uuid.New().String()
	tier := utils.SpinTierDB{
		ID: tierID,
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
		IsActive:        true,
	}
	if adminID != "" {
		tier.CreatedBy = &adminID
		tier.UpdatedBy = &adminID
	}

	if err := h.db.Create(&tier).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create tier",
			"details": err.Error(),
		})
		return
	}

	// Validate tier configuration after creation
	if err := h.calculator.ValidateTierConfiguration(); err != nil {
		// Rollback - delete the tier
		h.db.Delete(&tier)
		
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tier configuration validation failed",
			"details": err.Error(),
			"message": "The new tier creates overlapping or gapped tier ranges. Please adjust the amounts.",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Tier created successfully",
		"tier":    tier,
	})
}

// DeleteTier deletes a spin tier (soft delete - sets is_active to false)
// DELETE /api/admin/spin-tiers/:id
func (h *AdminSpinTiersHandler) DeleteTier(c *gin.Context) {
	tierID := c.Param("id")

	// Check if tier exists
	tier, err := h.calculator.GetTierByIDFromDB(tierID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Tier not found",
			"details": err.Error(),
		})
		return
	}

	// Soft delete - set is_active to false
	if err := h.db.Model(&utils.SpinTierDB{}).Where("id = ?", tierID).Update("is_active", false).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete tier",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tier deleted successfully",
		"tier":    tier,
	})
}

// ValidateTierConfiguration validates the current tier configuration
// GET /api/admin/spin-tiers/validate
func (h *AdminSpinTiersHandler) ValidateTierConfiguration(c *gin.Context) {
	if err := h.calculator.ValidateTierConfiguration(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"valid":   false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"valid":   true,
		"message": "Tier configuration is valid",
	})
}
