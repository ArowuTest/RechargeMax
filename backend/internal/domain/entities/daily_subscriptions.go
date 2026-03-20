package entities

import (
	"time"

	"github.com/google/uuid"
)

// DailySubscription is a user's recurring daily subscription plan.
//
// One row = one subscription "line".  A user can have multiple active lines
// simultaneously.  e.g.:
//   Line 1: bundle_quantity=1  daily_amount=2000  (₦20  → 1 entry/day)
//   Line 2: bundle_quantity=10 daily_amount=20000 (₦200 → 10 entries/day)
//   Line 3: bundle_quantity=5  daily_amount=10000 (₦100 → 5 entries/day)
//   Total:  16 entries/day and 16 points/day until each line is cancelled.
//
// Billing lifecycle:
//   pending  → payment link sent, waiting for first Paystack confirmation
//   active   → first payment confirmed; paystack_authorization_code stored;
//              auto-renews daily via SubscriptionBillingJob
//   paused   → billing failed for consecutive_failure_limit days in a row;
//              user must manually resume or re-subscribe
//   cancelled → user-cancelled; no further billing attempts
//   expired  → one-off subscription (auto_renew=false) that has run its course
//
// Maps to the daily_subscriptions table.
type DailySubscription struct {
	ID                        uuid.UUID  `json:"id"                          gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	SubscriptionCode          string     `json:"subscription_code"           gorm:"column:subscription_code;uniqueIndex;size:50"`
	UserID                    *uuid.UUID `json:"user_id"                     gorm:"column:user_id;type:uuid;index"`
	MSISDN                    string     `json:"msisdn"                      gorm:"column:msisdn;not null;index"`
	TierID                    *uuid.UUID `json:"tier_id"                     gorm:"column:tier_id;type:uuid"`
	// BundleQuantity is the number of draw entries (and points) awarded per successful billing day.
	// Pricing: ₦20 per entry → bundle_quantity=5 costs ₦100/day.
	BundleQuantity            int        `json:"bundle_quantity"             gorm:"column:bundle_quantity;not null;default:1"`
	TotalEntries              int        `json:"total_entries"               gorm:"column:total_entries;not null;default:0"`
	// DailyAmount in kobo (₦20 = 2000 kobo, ₦200 = 20000 kobo, etc.)
	DailyAmount               int64      `json:"daily_amount"                gorm:"column:daily_amount;not null;default:2000"`
	Amount                    int64      `json:"amount"                      gorm:"column:amount;type:bigint"` // legacy compat
	DrawEntriesEarned         *int       `json:"draw_entries_earned"         gorm:"column:draw_entries_earned;default:1"`
	PointsEarned              *int       `json:"points_earned"               gorm:"column:points_earned;default:0"`
	// Lifetime totals (updated by billing job after each successful day)
	TotalBilledAmount         int64      `json:"total_billed_amount"         gorm:"column:total_billed_amount;not null;default:0"`
	TotalPointsAwarded        int        `json:"total_points_awarded"        gorm:"column:total_points_awarded;not null;default:0"`
	Status                    string     `json:"status"                      gorm:"column:status;default:'pending'"`
	AutoRenew                 bool       `json:"auto_renew"                  gorm:"column:auto_renew;default:true"`
	IsPaid                    *bool      `json:"is_paid"                     gorm:"column:is_paid;default:false"`
	// Recurring billing fields
	NextBillingDate           time.Time  `json:"next_billing_date"           gorm:"column:next_billing_date"`
	LastBillingDate           *time.Time `json:"last_billing_date"           gorm:"column:last_billing_date"`
	// ConsecutiveFailures is incremented every time a full day's retries are exhausted.
	// When it reaches ConsecutiveFailureLimit the subscription is auto-paused.
	ConsecutiveFailures       int        `json:"consecutive_failures"        gorm:"column:consecutive_failures;not null;default:0"`
	PaymentMethod             string     `json:"payment_method"              gorm:"column:payment_method;not null;default:'paystack'"`
	PaymentReference          *string    `json:"payment_reference"           gorm:"column:payment_reference"`
	// PaystackAuthorizationCode is the reusable token stored after the FIRST successful
	// Paystack charge.  All subsequent daily auto-charges use this token via the
	// Paystack "charge authorization" API — the user never needs to re-enter card details.
	PaystackAuthorizationCode *string    `json:"paystack_authorization_code" gorm:"column:paystack_authorization_code"`
	PaystackCustomerCode      *string    `json:"paystack_customer_code"      gorm:"column:paystack_customer_code"`
	SubscriptionDate          time.Time  `json:"subscription_date"           gorm:"column:subscription_date;not null"`
	CancelledAt               *time.Time `json:"cancelled_at"                gorm:"column:cancelled_at"`
	CancellationReason        string     `json:"cancellation_reason"         gorm:"column:cancellation_reason"`
	PausedAt                  *time.Time `json:"paused_at"                   gorm:"column:paused_at"`
	CustomerEmail             *string    `json:"customer_email"              gorm:"column:customer_email"`
	CustomerName              *string    `json:"customer_name"               gorm:"column:customer_name"`
	CreatedAt                 time.Time  `json:"created_at"                  gorm:"column:created_at;autoCreateTime"`
	UpdatedAt                 time.Time  `json:"updated_at"                  gorm:"column:updated_at;autoUpdateTime"`

	// Virtual — populated on read, not stored
	Billings []SubscriptionBilling `json:"billings,omitempty" gorm:"foreignKey:SubscriptionID"`
}

// TableName maps to the daily_subscriptions table.
func (DailySubscription) TableName() string { return "daily_subscriptions" }

// ConsecutiveFailureLimit is the number of consecutive daily billing failures
// that will auto-pause a subscription.
const ConsecutiveFailureLimit = 7

// PricePerEntry is the canonical cost of one draw entry in kobo (₦20).
const PricePerEntry int64 = 2000
