package services

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"rechargemax/internal/domain/entities"
)

// ─────────────────────────────────────────────────────────────────────────────
// PlatformSettingsService
// ─────────────────────────────────────────────────────────────────────────────

// PlatformSettingsService manages application-wide key–value settings.
type PlatformSettingsService struct {
	db *gorm.DB
}

// NewPlatformSettingsService constructs a PlatformSettingsService.
func NewPlatformSettingsService(db *gorm.DB) *PlatformSettingsService {
	return &PlatformSettingsService{db: db}
}

// Upsert inserts or updates a single setting.
func (s *PlatformSettingsService) Upsert(_ context.Context, key, value, description string) error {
	return s.db.Exec(
		`INSERT INTO platform_settings (setting_key, setting_value, description)
		 VALUES (?, ?, ?)
		 ON CONFLICT (setting_key) DO UPDATE SET setting_value = EXCLUDED.setting_value, updated_at = now()`,
		key, value, description,
	).Error
}

// GetByKey returns a single setting by key, or an error if not found.
func (s *PlatformSettingsService) GetByKey(_ context.Context, key string) (*entities.PlatformSetting, error) {
	var setting entities.PlatformSetting
	if err := s.db.Where("setting_key = ?", key).First(&setting).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}

// LoadAll returns every setting as a key→value map.
func (s *PlatformSettingsService) LoadAll(_ context.Context) (map[string]string, error) {
	var rows []entities.PlatformSetting
	if err := s.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[string]string, len(rows))
	for _, r := range rows {
		result[r.SettingKey] = r.SettingValue
	}
	return result, nil
}

// UpsertMap upserts a batch of settings at once.  key is the full dotted key.
func (s *PlatformSettingsService) UpsertMap(ctx context.Context, settings map[string]interface{}) error {
	for key, val := range settings {
		if err := s.Upsert(ctx, key, fmt.Sprintf("%v", val), ""); err != nil {
			return fmt.Errorf("failed to save setting %q: %w", key, err)
		}
	}
	return nil
}

// UpsertCategory upserts all fields under `<category>.<field>` keys.
func (s *PlatformSettingsService) UpsertCategory(ctx context.Context, category string, fields map[string]interface{}) error {
	for field, val := range fields {
		key := category + "." + field
		if err := s.Upsert(ctx, key, fmt.Sprintf("%v", val), ""); err != nil {
			return fmt.Errorf("failed to save setting %q: %w", key, err)
		}
	}
	return nil
}

// BuildSettingsMap merges DB values over hardcoded defaults and groups by prefix.
// This is a pure function — no DB calls — so it can live on the service for reuse.
func BuildSettingsMap(dbValues map[string]string) map[string]interface{} {
	merged := make(map[string]string)
	for k, v := range DefaultSettings {
		merged[k] = v
	}
	for k, v := range dbValues {
		switch k {
		case "spin_wheel_enabled":
			merged["features.spin_wheel_enabled"] = v
		case "spin_wheel_minimum":
			merged["spin.min_recharge_amount"] = v
		default:
			merged[k] = v
		}
	}

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
		switch val {
		case "true":
			categories[cat][field] = true
		case "false":
			categories[cat][field] = false
		default:
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

// DefaultSettings provides fallback values when a key is absent from the DB.
var DefaultSettings = map[string]string{
	"platform.name":                        "RechargeMax Rewards",
	"platform.description":                 "Gamified mobile recharge platform with rewards",
	"platform.version":                     "1.0.0",
	"platform.environment":                 "development",
	"branding.logo_url":                    "",
	"branding.primary_color":               "#4F46E5",
	"branding.company_name":                "RechargeMax",
	"branding.support_email":               "support@rechargemax.ng",
	"branding.support_phone":               "+234-XXX-XXX-XXXX",
	"features.daily_draw_enabled":          "true",
	"features.spin_wheel_enabled":          "true",
	"features.affiliates_enabled":          "true",
	"features.ussd_enabled":                "true",
	"features.daily_subscription_enabled":  "true",
	"points.points_per_naira":              "0.005",
	"points.min_recharge_for_points":       "200",
	"spin.enabled":                         "true",
	"spin.min_recharge_amount":             "1000",
	"spin.daily_spin_limit":                "10",
	"spin.daily_spins":                     "3",
	"daily_subscription.daily_price":       "20",
	"daily_subscription.weekly_price":      "100",
	"daily_subscription.monthly_price":     "300",
	"daily_subscription.daily_spins":       "3",
	"daily_subscription.weekly_spins":      "25",
	"daily_subscription.monthly_spins":     "100",
	"daily_subscription.auto_renewal":      "true",
	"daily_subscription.grace_period_days": "3",
	"recharge.provider":                    "vtpass",
	"recharge.mode":                        "sandbox",
	"recharge.min_amount":                  "50",
	"recharge.max_amount":                  "50000",
	"recharge.commission_rate":             "2.5",
	"security.jwt_expiry_hours":            "24",
	"security.admin_jwt_expiry_hours":      "8",
	"security.password_min_length":         "8",
	"security.max_login_attempts":          "5",
	"security.lockout_duration_minutes":    "30",
	"notifications.sms_enabled":            "true",
	"notifications.email_enabled":          "true",
	"notifications.sms_provider":           "termii",
	"notifications.sms_sender_id":          "RechargeMax",
}
