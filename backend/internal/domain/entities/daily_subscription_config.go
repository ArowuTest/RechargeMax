package entities

import (
	"time"

	"github.com/google/uuid"
)

// DailySubscriptionConfig represents the daily_subscription_config table.
// Admin-configurable settings for the daily subscription product.
type DailySubscriptionConfig struct {
	ID                 uuid.UUID  `json:"id" gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	Amount             float64    `json:"amount" gorm:"column:amount;type:numeric(5,2);not null" validate:"required,gt=0"` // Naira — numeric(5,2) in DB, must be float64
	DrawEntriesEarned  *int       `json:"draw_entries_earned" gorm:"column:draw_entries_earned"`
	IsPaid             *bool      `json:"is_paid" gorm:"column:is_paid"`
	Description        string     `json:"description" gorm:"column:description"`
	TermsAndConditions string     `json:"terms_and_conditions" gorm:"column:terms_and_conditions"`
	CreatedAt          *time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt          *time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

func (DailySubscriptionConfig) TableName() string {
	return "daily_subscription_config"
}
