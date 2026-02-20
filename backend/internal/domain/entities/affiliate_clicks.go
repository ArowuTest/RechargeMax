package entities

import (
	"time"

)

// AffiliateClicks represents the affiliate_clicks table
type AffiliateClicks struct {
	IpAddress   string     `json:"ip_address" gorm:"column:ip_address"`
	UserAgent   string     `json:"user_agent" gorm:"column:user_agent"`
	ReferrerUrl string     `json:"referrer_url" gorm:"column:referrer_url"`
	LandingPage string     `json:"landing_page" gorm:"column:landing_page"`
	Converted   *bool      `json:"converted" gorm:"column:converted"`
	CreatedAt   *time.Time `json:"created_at" gorm:"column:created_at"`
	ConvertedAt *time.Time `json:"converted_at" gorm:"column:converted_at"`
}

// TableName specifies the table name for AffiliateClicks
func (AffiliateClicks) TableName() string {
	return "affiliate_clicks"
}
