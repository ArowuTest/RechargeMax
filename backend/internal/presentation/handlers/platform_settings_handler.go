package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"rechargemax/internal/application/services"
)

// PlatformSettingsHandler handles admin platform-settings CRUD.
type PlatformSettingsHandler struct {
	settingsSvc *services.PlatformSettingsService
}

// NewPlatformSettingsHandler creates a new PlatformSettingsHandler.
func NewPlatformSettingsHandler(settingsSvc *services.PlatformSettingsService) *PlatformSettingsHandler {
	return &PlatformSettingsHandler{settingsSvc: settingsSvc}
}

// GetAllSettings returns all settings merged with in-code defaults, grouped by category.
func (h *PlatformSettingsHandler) GetAllSettings(c *gin.Context) {
	dbValues, _ := h.settingsSvc.LoadAll(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    services.BuildSettingsMap(dbValues),
	})
}

// GetSettingsByCategory returns settings for a single category prefix.
func (h *PlatformSettingsHandler) GetSettingsByCategory(c *gin.Context) {
	dbValues, _ := h.settingsSvc.LoadAll(c.Request.Context())
	all := services.BuildSettingsMap(dbValues)
	category := c.Param("category")
	settings, ok := all[category]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Category not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": settings})
}

// UpdateSettings upserts multiple settings from a flat key→value map.
func (h *PlatformSettingsHandler) UpdateSettings(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request data"})
		return
	}
	if err := h.settingsSvc.UpsertMap(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Settings updated successfully", "data": req})
}

// UpdateCategorySettings upserts all fields under the given category prefix.
func (h *PlatformSettingsHandler) UpdateCategorySettings(c *gin.Context) {
	category := c.Param("category")
	valid := map[string]bool{
		"platform": true, "branding": true, "features": true, "points": true,
		"spin": true, "daily_subscription": true, "recharge": true,
		"security": true, "notifications": true,
	}
	if !valid[category] {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Invalid category"})
		return
	}
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request data"})
		return
	}
	if err := h.settingsSvc.UpsertCategory(c.Request.Context(), category, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Category settings updated successfully", "data": req})
}

// GetSetting returns a single setting by key.
func (h *PlatformSettingsHandler) GetSetting(c *gin.Context) {
	key := c.Param("key")
	setting, err := h.settingsSvc.GetByKey(c.Request.Context(), key)
	if err != nil {
		if defVal, ok := services.DefaultSettings[key]; ok {
			c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"key": key, "value": defVal}})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Setting not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"key": setting.SettingKey, "value": setting.SettingValue, "description": setting.Description},
	})
}

// UpdateSetting upserts a single setting by key.
func (h *PlatformSettingsHandler) UpdateSetting(c *gin.Context) {
	key := c.Param("key")
	var req struct {
		Value       string `json:"value" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request data"})
		return
	}
	if err := h.settingsSvc.Upsert(c.Request.Context(), key, req.Value, req.Description); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to update setting"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Setting updated successfully", "data": gin.H{"key": key, "value": req.Value}})
}
