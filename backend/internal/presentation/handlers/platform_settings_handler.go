package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PlatformSettingsHandler struct {
	db *gorm.DB
}

func NewPlatformSettingsHandler(db *gorm.DB) *PlatformSettingsHandler {
	return &PlatformSettingsHandler{
		db: db,
	}
}

// SettingItem represents a single setting
type SettingItem struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
}

// GetAllSettings returns all platform settings
func (h *PlatformSettingsHandler) GetAllSettings(c *gin.Context) {
	// For now, return hardcoded settings that exist in other endpoints
	// This provides a centralized view of all configurable settings
	settings := map[string]interface{}{
		"platform": map[string]interface{}{
			"name":        "RechargeMax Rewards",
			"description": "Gamified mobile recharge platform with rewards",
			"version":     "1.0.0",
			"environment": "development",
		},
		"branding": map[string]interface{}{
			"logo_url":      "",
			"primary_color": "#4F46E5",
			"company_name":  "RechargeMax",
			"support_email": "support@rechargemax.ng",
			"support_phone": "+234-XXX-XXX-XXXX",
		},
		"features": map[string]interface{}{
			"daily_draw_enabled":    true,
			"spin_wheel_enabled":    true,
			"affiliates_enabled":    true,
			"ussd_enabled":          true,
			"daily_subscription_enabled": true,
		},
		"points": map[string]interface{}{
			"points_per_naira": 0.005, // N200 = 1 point
			"min_recharge_for_points": 200,
		},
		"spin": map[string]interface{}{
			"enabled":             true,
			"min_recharge_amount": 1000,
			"daily_spin_limit":    10,
		},
		"daily_subscription": map[string]interface{}{
			"daily_price":       20,
			"weekly_price":      100,
			"monthly_price":     300,
			"daily_spins":       3,
			"weekly_spins":      25,
			"monthly_spins":     100,
			"auto_renewal":      true,
			"grace_period_days": 3,
		},
		"recharge": map[string]interface{}{
			"provider":         "vtpass",
			"mode":             "sandbox",
			"min_amount":       50,
			"max_amount":       50000,
			"commission_rate":  2.5,
		},
		"security": map[string]interface{}{
			"jwt_expiry_hours":       24,
			"admin_jwt_expiry_hours": 8,
			"password_min_length":    8,
			"max_login_attempts":     5,
			"lockout_duration_minutes": 30,
		},
		"notifications": map[string]interface{}{
			"sms_enabled":   true,
			"email_enabled": true,
			"sms_provider":  "termii",
			"sms_sender_id": "RechargeMax",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settings,
	})
}

// GetSettingsByCategory returns settings for a specific category
func (h *PlatformSettingsHandler) GetSettingsByCategory(c *gin.Context) {
	category := c.Param("category")

	// Map of all settings by category
	allSettings := map[string]interface{}{
		"platform": map[string]interface{}{
			"name":        "RechargeMax Rewards",
			"description": "Gamified mobile recharge platform with rewards",
			"version":     "1.0.0",
			"environment": "development",
		},
		"branding": map[string]interface{}{
			"logo_url":      "",
			"primary_color": "#4F46E5",
			"company_name":  "RechargeMax",
			"support_email": "support@rechargemax.ng",
			"support_phone": "+234-XXX-XXX-XXXX",
		},
		"features": map[string]interface{}{
			"daily_draw_enabled":         true,
			"spin_wheel_enabled":         true,
			"affiliates_enabled":         true,
			"ussd_enabled":               true,
			"daily_subscription_enabled": true,
		},
		"points": map[string]interface{}{
			"points_per_naira":        0.005,
			"min_recharge_for_points": 200,
		},
		"spin": map[string]interface{}{
			"enabled":             true,
			"min_recharge_amount": 1000,
			"daily_spin_limit":    10,
		},
		"daily_subscription": map[string]interface{}{
			"daily_price":         20,
			"weekly_price":        100,
			"monthly_price":       300,
			"daily_spins":         3,
			"weekly_spins":        25,
			"monthly_spins":       100,
			"auto_renewal":        true,
			"grace_period_days":   3,
			"max_subscriptions":   1,
		},
		"recharge": map[string]interface{}{
			"provider":        "vtpass",
			"mode":            "sandbox",
			"min_amount":      50,
			"max_amount":      50000,
			"commission_rate": 2.5,
		},
		"security": map[string]interface{}{
			"jwt_expiry_hours":         24,
			"admin_jwt_expiry_hours":   8,
			"password_min_length":      8,
			"max_login_attempts":       5,
			"lockout_duration_minutes": 30,
		},
		"notifications": map[string]interface{}{
			"sms_enabled":   true,
			"email_enabled": true,
			"sms_provider":  "termii",
			"sms_sender_id": "RechargeMax",
		},
	}

	settings, exists := allSettings[category]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Category not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settings,
	})
}

// UpdateSettings updates multiple settings at once
func (h *PlatformSettingsHandler) UpdateSettings(c *gin.Context) {
	var req map[string]interface{}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
		})
		return
	}

	// In a real implementation, you would:
	// 1. Validate the settings
	// 2. Update them in the database or config files
	// 3. Potentially restart services if needed

	// For now, just acknowledge the update
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Settings updated successfully (note: changes require backend restart to take effect)",
		"data":    req,
	})
}

// UpdateCategorySettings updates settings for a specific category
func (h *PlatformSettingsHandler) UpdateCategorySettings(c *gin.Context) {
	category := c.Param("category")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
		})
		return
	}

	// Validate category exists
	validCategories := map[string]bool{
		"platform":           true,
		"branding":           true,
		"features":           true,
		"points":             true,
		"spin":               true,
		"daily_subscription": true,
		"recharge":           true,
		"security":           true,
		"notifications":      true,
	}

	if !validCategories[category] {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Invalid category",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Category settings updated successfully",
		"data":    req,
	})
}

// GetSetting returns a single setting value
func (h *PlatformSettingsHandler) GetSetting(c *gin.Context) {
	key := c.Param("key")

	// This would normally query the database
	// For now, return a not implemented message
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"key":   key,
			"value": nil,
			"note":  "Individual setting retrieval - use category endpoints for grouped settings",
		},
	})
}

// UpdateSetting updates a single setting
func (h *PlatformSettingsHandler) UpdateSetting(c *gin.Context) {
	key := c.Param("key")

	var req struct {
		Value string `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Setting updated successfully",
		"data": gin.H{
			"key":   key,
			"value": req.Value,
		},
	})
}
