package entities

import (
	"time"

)

// AffiliateAnalytics represents the affiliate_analytics table
type AffiliateAnalytics struct {
	AnalyticsDate           time.Time  `json:"analytics_date" gorm:"column:analytics_date;not null" validate:"required"`
	TotalClicks             *int       `json:"total_clicks" gorm:"column:total_clicks"`
	UniqueClicks            *int       `json:"unique_clicks" gorm:"column:unique_clicks;uniqueIndex"`
	Conversions             *int       `json:"conversions" gorm:"column:conversions"`
	ConversionRate          *float64   `json:"conversion_rate" gorm:"column:conversion_rate"`
	TotalCommission         *float64   `json:"total_commission" gorm:"column:total_commission"`
	RechargeCommissions     *float64   `json:"recharge_commissions" gorm:"column:recharge_commissions"`
	SubscriptionCommissions *float64   `json:"subscription_commissions" gorm:"column:subscription_commissions"`
	TopReferrerCountry      string     `json:"top_referrer_country" gorm:"column:top_referrer_country"`
	TopDeviceType           string     `json:"top_device_type" gorm:"column:top_device_type"`
	CreatedAt               *time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt               *time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// TableName specifies the table name for AffiliateAnalytics
func (AffiliateAnalytics) TableName() string {
	return "affiliate_analytics"
}
