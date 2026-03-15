package entities

import (
	"time"

	"github.com/google/uuid"
)

// DailySubscription represents a user's active daily subscription.
// Maps to the daily_subscriptions table.
type DailySubscription struct {
	ID                 uuid.UUID  `json:"id"                  gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	SubscriptionCode   string     `json:"subscription_code"   gorm:"column:subscription_code;uniqueIndex;size:50"`
	UserID             *uuid.UUID `json:"user_id"             gorm:"column:user_id;type:uuid;index"`
	MSISDN             string     `json:"msisdn"              gorm:"column:msisdn;not null;index"                  validate:"required"`
	TierID             uuid.UUID  `json:"tier_id"             gorm:"column:tier_id;type:uuid;not null"              validate:"required"`
	BundleQuantity     int        `json:"bundle_quantity"     gorm:"column:bundle_quantity;not null;default:1"      validate:"required,min=1"`
	TotalEntries       int        `json:"total_entries"       gorm:"column:total_entries;not null"`
	DailyAmount        int64      `json:"daily_amount"        gorm:"column:daily_amount;not null"`
	DrawEntriesEarned  *int       `json:"draw_entries_earned" gorm:"column:draw_entries_earned;default:1"`
	PointsEarned       *int       `json:"points_earned"       gorm:"column:points_earned;default:0"`
	Status             string     `json:"status"              gorm:"column:status;default:'active'"                 validate:"oneof=active paused cancelled"`
	AutoRenew          bool       `json:"auto_renew"          gorm:"column:auto_renew;default:true"`
	IsPaid             *bool      `json:"is_paid"             gorm:"column:is_paid;default:false"`
	NextBillingDate    time.Time  `json:"next_billing_date"   gorm:"column:next_billing_date;not null"`
	LastBillingDate    *time.Time `json:"last_billing_date"   gorm:"column:last_billing_date"`
	PaymentMethod      string     `json:"payment_method"      gorm:"column:payment_method;not null"`
	PaymentReference   *string    `json:"payment_reference"   gorm:"column:payment_reference"`
	SubscriptionDate   time.Time  `json:"subscription_date"   gorm:"column:subscription_date;not null"`
	CancelledAt        *time.Time `json:"cancelled_at"        gorm:"column:cancelled_at"`
	CancellationReason string     `json:"cancellation_reason" gorm:"column:cancellation_reason"`
	CustomerEmail      *string    `json:"customer_email"      gorm:"column:customer_email"`
	CustomerName       *string    `json:"customer_name"       gorm:"column:customer_name"`
	Amount             int64      `json:"amount"              gorm:"column:amount;type:bigint"` // legacy compat column
	CreatedAt          time.Time  `json:"created_at"          gorm:"column:created_at;autoCreateTime"`
	UpdatedAt          time.Time  `json:"updated_at"          gorm:"column:updated_at;autoUpdateTime"`
}

// TableName maps to the daily_subscriptions table.
func (DailySubscription) TableName() string {
	return "daily_subscriptions"
}
