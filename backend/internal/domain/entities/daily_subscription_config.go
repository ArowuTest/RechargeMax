package entities

import (
	"time"

)

// DailySubscriptionConfig represents the daily_subscription_config table
type DailySubscriptionConfig struct {
	Amount             int64      `json:"amount" gorm:"column:amount;type:bigint;not null" validate:"required"`
	DrawEntriesEarned  *int       `json:"draw_entries_earned" gorm:"column:draw_entries_earned"`
	IsPaid             *bool      `json:"is_paid" gorm:"column:is_paid"`
	Description        string     `json:"description" gorm:"column:description"`
	TermsAndConditions string     `json:"terms_and_conditions" gorm:"column:terms_and_conditions"`
	CreatedAt          *time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt          *time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// TableName specifies the table name for DailySubscriptionConfig
func (DailySubscriptionConfig) TableName() string {
	return "daily_subscription_config"
}
