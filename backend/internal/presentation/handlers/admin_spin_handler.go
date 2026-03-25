package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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

	// Persist each config field via PlatformSettingsService ("spin.<field>" keys)
	if err := h.settingsSvc.UpsertCategory(c.Request.Context(), "spin", config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
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
		Value           float64  `json:"value"` // not required for NO_WIN type
		Probability     float64  `json:"probability" binding:"required"`
		IsActive        bool     `json:"is_active"`
		MinimumRecharge *float64 `json:"minimum_recharge"`
		ColorScheme     string   `json:"color_scheme"`
		Color           string   `json:"color"`
		SortOrder       *float64 `json:"sort_order"`
		IsNoWin         bool     `json:"is_no_win"`
		NoWinMessage    string   `json:"no_win_message"`
	}

	if err := c.ShouldBindJSON(&prizeData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// De-duplicate: if a prize with the same name, type, and value already exists, return it
	type existingPrize struct {
		ID string `gorm:"column:id"`
	}
	var existing existingPrize
	dupErr := h.db.WithContext(ctx).
		Table("wheel_prizes").
		Select("id").
		Where("prize_name = ? AND prize_type = ? AND prize_value = ?",
			prizeData.Name, strings.ToUpper(prizeData.Type), int64(prizeData.Value)).
		First(&existing).Error
	if dupErr == nil && existing.ID != "" {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Prize already exists (duplicate suppressed)",
			"data":    map[string]interface{}{"id": existing.ID},
		})
		return
	}

	prizeMap := map[string]interface{}{
		"name":          prizeData.Name,
		"type":          prizeData.Type,
		"value":         prizeData.Value,
		"probability":   prizeData.Probability,
		"is_active":     prizeData.IsActive,
		"is_no_win":     prizeData.IsNoWin,
		"no_win_message": prizeData.NoWinMessage,
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
		Name            *string  `json:"name"`
		Type            *string  `json:"type"`
		Value           *float64 `json:"value"`
		Probability     *float64 `json:"probability"`
		IsActive        *bool    `json:"is_active"`
		MinimumRecharge *float64 `json:"minimum_recharge"`
		ColorScheme     *string  `json:"color"`
		SortOrder       *int     `json:"sort_order"`
		IsNoWin         *bool    `json:"is_no_win"`
		NoWinMessage    *string  `json:"no_win_message"`
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
	if updateData.MinimumRecharge != nil {
		updateMap["minimum_recharge"] = *updateData.MinimumRecharge
	}
	if updateData.ColorScheme != nil {
		updateMap["color_scheme"] = *updateData.ColorScheme
		updateMap["color"] = *updateData.ColorScheme
	}
	if updateData.SortOrder != nil {
		updateMap["sort_order"] = float64(*updateData.SortOrder)
	}
	if updateData.IsNoWin != nil {
		updateMap["is_no_win"] = *updateData.IsNoWin
	}
	if updateData.NoWinMessage != nil {
		updateMap["no_win_message"] = *updateData.NoWinMessage
	}

	updatedPrize, err := h.spinService.UpdatePrize(ctx, prizeID.String(), updateMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(), // surface actual error to admin (probability overage, not found, etc.)
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

