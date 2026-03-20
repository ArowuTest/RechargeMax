package entities

import (
	"time"

	"github.com/google/uuid"
)

// SubscriptionBilling is the authoritative record of one daily billing attempt
// for one active subscription line.
//
// Lifecycle:
//   pending   → row created by billing job, charge not yet attempted
//   attempted → charge API called, waiting for Paystack webhook confirmation
//   completed → Paystack confirmed success; points_awarded set to true
//   failed    → all retries exhausted for this billing_date; points NOT awarded
//   skipped   → subscription was paused/cancelled before this day could be billed
//
// Retry schedule (controlled by billing job):
//   Attempt 0 (initial):   at next_billing_date  (e.g. 08:00)
//   Retry 1:               +1 hour
//   Retry 2:               +3 hours
//   Retry 3:               +8 hours
//   → If retry 3 also fails: status=failed, subscription.consecutive_failures++
//     next day's billing record is created as normal at next_billing_date+24h
//
// Points idempotency:
//   points_awarded=false on creation. Set to true atomically when the billing job
//   calls awardSubscriptionPoints().  Even if the webhook fires twice the award
//   is only executed once.
//
// Maps to the subscription_billings table.
type SubscriptionBilling struct {
	ID                     uuid.UUID  `json:"id"                       gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	SubscriptionID         uuid.UUID  `json:"subscription_id"          gorm:"column:subscription_id;not null;index"`
	MSISDN                 string     `json:"msisdn"                   gorm:"column:msisdn;not null;index"`
	// BillingDate is the calendar day (UTC date) that this billing covers.
	// Used for deduplication: UNIQUE(subscription_id, billing_date).
	BillingDate            time.Time  `json:"billing_date"             gorm:"column:billing_date;not null"`
	Amount                 int64      `json:"amount"                   gorm:"column:amount;not null"` // kobo
	EntriesToAward         int        `json:"entries_to_award"         gorm:"column:entries_to_award;not null;default:1"`
	PointsToAward          int        `json:"points_to_award"          gorm:"column:points_to_award;not null;default:1"`
	Status                 string     `json:"status"                   gorm:"column:status;not null;default:'pending'"`
	PaymentReference       *string    `json:"payment_reference"        gorm:"column:payment_reference"`
	PaystackTransactionID  *int64     `json:"paystack_transaction_id"  gorm:"column:paystack_transaction_id"`
	GatewayResponse        string     `json:"gateway_response"         gorm:"column:gateway_response"`
	FailureReason          string     `json:"failure_reason"           gorm:"column:failure_reason"`
	RetryCount             int        `json:"retry_count"              gorm:"column:retry_count;not null;default:0"`
	MaxRetries             int        `json:"max_retries"              gorm:"column:max_retries;not null;default:3"`
	NextRetryAt            *time.Time `json:"next_retry_at"            gorm:"column:next_retry_at"`
	// PointsAwarded is the idempotency flag — once true, points are never re-awarded.
	PointsAwarded          bool       `json:"points_awarded"           gorm:"column:points_awarded;not null;default:false"`
	ProcessedAt            *time.Time `json:"processed_at"             gorm:"column:processed_at"`
	CreatedAt              time.Time  `json:"created_at"               gorm:"column:created_at;autoCreateTime"`
	UpdatedAt              time.Time  `json:"updated_at"               gorm:"column:updated_at;autoUpdateTime"`
}

// TableName maps to the subscription_billings table.
func (SubscriptionBilling) TableName() string { return "subscription_billings" }

// RetryDelays defines the wait time before each retry attempt (index = retry_count after failure).
var RetryDelays = []time.Duration{
	1 * time.Hour,  // Retry 1: 1 hour after initial attempt
	3 * time.Hour,  // Retry 2: 3 hours after retry 1
	8 * time.Hour,  // Retry 3: 8 hours after retry 2
}
