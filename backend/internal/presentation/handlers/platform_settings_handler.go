package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

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

// PlatformSetting is the GORM model for the platform_settings table
type PlatformSetting struct {
	SettingKey   string    `gorm:"column:setting_key;primaryKey"`
	SettingValue string    `gorm:"column:setting_value"`
	Description  string    `gorm:"column:description"`
	IsPublic     bool      `gorm:"column:is_public"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (PlatformSetting) TableName() string {
	return "platform_settings"
}

// SettingItem represents a single setting
type SettingItem struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
}

// upsertSetting inserts or updates a single setting in the database
func (h *PlatformSettingsHandler) upsertSetting(key, value, description string) error {
	return h.db.Exec(
		`INSERT INTO platform_settings (setting_key, setting_value, description)
		 VALUES (?, ?, ?)
		 ON CONFLICT (setting_key) DO UPDATE SET setting_value = EXCLUDED.setting_value, updated_at = now()`,
		key, value, description,
	).Error
}

// loadSettingsFromDB loads all settings from the database into a map
func (h *PlatformSettingsHandler) loadSettingsFromDB() (map[string]string, error) {
	var rows []PlatformSetting
	if err := h.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[string]string, len(rows))
	for _, r := range rows {
		result[r.SettingKey] = r.SettingValue
	}
	return result, nil
}

// defaultSettings provides fallback values when a key is not in the DB
var defaultSettings = map[string]string{
	"platform.name":                    "RechargeMax Rewards",
	"platform.description":             "Gamified mobile recharge platform with rewards",
	"platform.version":                 "1.0.0",
	"platform.environment":             "development",
	"branding.logo_url":                "",
	"branding.primary_color":           "#4F46E5",
	"branding.company_name":            "RechargeMax",
	"branding.support_email":           "support@rechargemax.ng",
	"branding.support_phone":           "+234-XXX-XXX-XXXX",
	"features.daily_draw_enabled":      "true",
	"features.spin_wheel_enabled":      "true",
	"features.affiliates_enabled":      "true",
	"features.ussd_enabled":            "true",
	"features.daily_subscription_enabled": "true",
	"points.points_per_naira":          "0.005",
	"points.min_recharge_for_points":   "200",
	"spin.enabled":                     "true",
	"spin.min_recharge_amount":         "1000",
	"spin.daily_spin_limit":            "10",
	"spin.daily_spins":                 "3",
	"daily_subscription.daily_price":   "20",
	"daily_subscription.weekly_price":  "100",
	"daily_subscription.monthly_price": "300",
	"daily_subscription.daily_spins":   "3",
	"daily_subscription.weekly_spins":  "25",
	"daily_subscription.monthly_spins": "100",
	"daily_subscription.auto_renewal":  "true",
	"daily_subscription.grace_period_days": "3",
	"recharge.provider":                "vtpass",
	"recharge.mode":                    "sandbox",
	"recharge.min_amount":              "50",
	"recharge.max_amount":              "50000",
	"recharge.commission_rate":         "2.5",
	"security.jwt_expiry_hours":        "24",
	"security.admin_jwt_expiry_hours":  "8",
	"security.password_min_length":     "8",
	"security.max_login_attempts":      "5",
	"security.lockout_duration_minutes": "30",
	"notifications.sms_enabled":        "true",
	"notifications.email_enabled":      "true",
	"notifications.sms_provider":       "termii",
	"notifications.sms_sender_id":      "RechargeMax",
}

// buildSettingsMap merges DB values over defaults and groups by category
func buildSettingsMap(dbValues map[string]string) map[string]interface{} {
	merged := make(map[string]string)
	// Start with defaults
	for k, v := range defaultSettings {
		merged[k] = v
	}
	// Override with DB values (DB keys may be flat like "spin_wheel_enabled" or dotted)
	for k, v := range dbValues {
		// Normalise legacy flat keys to dotted form
		switch k {
		case "spin_wheel_enabled":
			merged["features.spin_wheel_enabled"] = v
		case "spin_wheel_minimum":
			merged["spin.min_recharge_amount"] = v
		default:
			merged[k] = v
		}
	}

	// Group by category prefix
	categories := make(map[string]map[string]interface{})
	for key, val := range merged {
		parts := strings.SplitN(key, ".", 2)
		if len(parts) != 2 {
			continue
		}
		cat, field := parts[0], parts[1]
		if categories[cat] == nil {
			categories[cat] = make(map[string]interface{})
		}
		// Try to parse booleans and numbers
		switch val {
		case "true":
			categories[cat][field] = true
		case "false":
			categories[cat][field] = false
		default:
			// Try numeric
			var f float64
			if _, err := fmt.Sscanf(val, "%f", &f); err == nil {
				categories[cat][field] = f
			} else {
				categories[cat][field] = val
			}
		}
	}
	result := make(map[string]interface{})
	for k, v := range categories {
		result[k] = v
	}
	return result
}

// GetAllSettings returns all platform settings, reading from DB with fallback to defaults
func (h *PlatformSettingsHandler) GetAllSettings(c *gin.Context) {
	dbValues, err := h.loadSettingsFromDB()
	if err != nil {
		// Fall back to defaults if DB is unavailable
		dbValues = make(map[string]string)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    buildSettingsMap(dbValues),
	})
}

// GetSettingsByCategory returns settings for a specific category
func (h *PlatformSettingsHandler) GetSettingsByCategory(c *gin.Context) {
	category := c.Param("category")

	dbValues, err := h.loadSettingsFromDB()
	if err != nil {
		dbValues = make(map[string]string)
	}

	all := buildSettingsMap(dbValues)
	settings, exists := all[category]
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

// UpdateSettings updates multiple settings at once (flat key=value map)
func (h *PlatformSettingsHandler) UpdateSettings(c *gin.Context) {
	var req map[string]interface{}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
		})
		return
	}

	for key, val := range req {
		strVal := fmt.Sprintf("%v", val)
		if err := h.upsertSetting(key, strVal, ""); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to save setting: " + key,
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Settings updated successfully",
		"data":    req,
	})
}

// UpdateCategorySettings updates settings for a specific category (stores as category.field keys)
func (h *PlatformSettingsHandler) UpdateCategorySettings(c *gin.Context) {
	category := c.Param("category")

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

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
		})
		return
	}

	for field, val := range req {
		key := category + "." + field
		strVal := fmt.Sprintf("%v", val)
		if err := h.upsertSetting(key, strVal, ""); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to save setting: " + key,
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Category settings updated successfully",
		"data":    req,
	})
}

// GetSetting returns a single setting value by key
func (h *PlatformSettingsHandler) GetSetting(c *gin.Context) {
	key := c.Param("key")

	var setting PlatformSetting
	err := h.db.Where("setting_key = ?", key).First(&setting).Error
	if err != nil {
		// Try default
		if defVal, ok := defaultSettings[key]; ok {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": gin.H{
					"key":   key,
					"value": defVal,
				},
			})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Setting not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"key":         setting.SettingKey,
			"value":       setting.SettingValue,
			"description": setting.Description,
		},
	})
}

// UpdateSetting updates a single setting by key
func (h *PlatformSettingsHandler) UpdateSetting(c *gin.Context) {
	key := c.Param("key")

	var req struct {
		Value       string `json:"value" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
		})
		return
	}

	if err := h.upsertSetting(key, req.Value, req.Description); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update setting",
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
