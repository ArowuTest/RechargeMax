package handlers

import (
	"fmt"
	"net/http"
	"strings"

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
// Also injects any flat (no-dot) DB keys directly into the nested map under a "_flat" key
// so the frontend can discover keys like "registration_enabled" without a category prefix.
func (h *PlatformSettingsHandler) GetAllSettings(c *gin.Context) {
	dbValues, _ := h.settingsSvc.LoadAll(c.Request.Context())
	nested := services.BuildSettingsMap(dbValues)

	// Inject flat keys (keys without a ".") into the nested map under their own name so
	// the frontend flat-key layout schema (e.g. "registration_enabled") can find them.
	// We create a pseudo-category for each flat key using an empty prefix trick:
	// actually we just copy them into the nested map directly as top-level entries.
	flatGroup := make(map[string]interface{})
	for k, v := range dbValues {
		if !strings.Contains(k, ".") {
			switch v {
			case "true":
				flatGroup[k] = true
			case "false":
				flatGroup[k] = false
			default:
				var f float64
				if _, err := fmt.Sscanf(v, "%f", &f); err == nil {
					flatGroup[k] = f
				} else {
					flatGroup[k] = v
				}
			}
		}
	}
	// Also inject flat keys from DefaultSettings that aren't already present
	for k, v := range services.DefaultSettings {
		if !strings.Contains(k, ".") {
			if _, exists := flatGroup[k]; !exists {
				flatGroup[k] = v
			}
		}
	}
	// Merge into nested: each flat key becomes its own top-level entry
	// Frontend's nested parser: for [cat, items] — if items is a string/bool/number, skip
	// So we use a synthetic category for flat keys
	if len(flatGroup) > 0 {
		nested["_flat"] = flatGroup
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    nested,
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
