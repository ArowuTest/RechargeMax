package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"rechargemax/internal/domain/entities"
)

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
		"config":               config,
		"daily_revenue":        0, // Would calculate from billing records
		"monthly_revenue":      0, // Would calculate from billing records
		"churn_rate":           0, // Would calculate from cancellations
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
		drawTypeID, parseErr := uuid.Parse(drawTypeIDStr)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid draw_type_id",
			})
			return
		}
		templates, err = h.prizeTemplateService.GetTemplatesByDrawType(drawTypeID)
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
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid template ID",
		})
		return
	}

	template, err := h.prizeTemplateService.GetTemplate(templateID)
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
		"success":          true,
		"data":             template,
		"total_prize_pool": totalPool,
	})
}

// CreatePrizeTemplate creates a new prize template with categories
func (h *AdminComprehensiveHandler) CreatePrizeTemplate(c *gin.Context) {
	var req struct {
		Name        string                   `json:"name" binding:"required"`
		Description string                   `json:"description"`
		DrawTypeID  uuid.UUID                `json:"draw_type_id" binding:"required"`
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
	templateID, err := uuid.Parse(templateIDStr)
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
		templateID,
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
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid template ID",
		})
		return
	}

	if err := h.prizeTemplateService.DeleteTemplate(templateID); err != nil {
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
	templateID, err := uuid.Parse(templateIDStr)
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
		templateID,
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
	categoryID, err := uuid.Parse(categoryIDStr)
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
		categoryID,
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
	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid category ID",
		})
		return
	}

	if err := h.prizeTemplateService.DeleteCategory(categoryID); err != nil {
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
