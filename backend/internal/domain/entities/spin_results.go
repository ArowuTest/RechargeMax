package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SpinResult represents the spin_results table
type SpinResult struct {
	ID            uuid.UUID  `json:"id" gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	SpinCode      string     `json:"spin_code" gorm:"column:spin_code;uniqueIndex;size:30"`
	UserID        *uuid.UUID `json:"user_id" gorm:"column:user_id;index"`
	TransactionID *uuid.UUID `json:"transaction_id" gorm:"column:transaction_id"`

	// Spin details
	Msisdn     string     `json:"msisdn" gorm:"column:msisdn;not null;index" validate:"required"`
	PrizeID    *uuid.UUID `json:"prize_id" gorm:"column:prize_id"` // Reference to wheel_prizes
	PrizeName  string     `json:"prize_name" gorm:"column:prize_name;not null" validate:"required"`
	PrizeType  string     `json:"prize_type" gorm:"column:prize_type;not null" validate:"required"`
	PrizeValue int64      `json:"prize_value" gorm:"column:prize_value;type:bigint;not null" validate:"required"` // Value in kobo

	// Claim status
	ClaimStatus    string     `json:"claim_status" gorm:"column:claim_status;default:PENDING;check:claim_status IN ('PENDING','CLAIMED','EXPIRED','PENDING_ADMIN_REVIEW','APPROVED','REJECTED')" validate:"oneof=PENDING CLAIMED EXPIRED PENDING_ADMIN_REVIEW APPROVED REJECTED"`
	ClaimedAt      *time.Time `json:"claimed_at" gorm:"column:claimed_at"`
	ClaimReference string     `json:"claim_reference" gorm:"column:claim_reference"`

	// Admin review fields (for CASH prizes requiring approval)
	ReviewedBy       *uuid.UUID `json:"reviewed_by" gorm:"column:reviewed_by"`
	ReviewedAt       *time.Time `json:"reviewed_at" gorm:"column:reviewed_at"`
	RejectionReason  string     `json:"rejection_reason" gorm:"column:rejection_reason"`
	AdminNotes       string     `json:"admin_notes" gorm:"column:admin_notes"`
	PaymentReference string     `json:"payment_reference" gorm:"column:payment_reference"`

	// Bank details (for CASH prizes)
	BankAccountNumber string `json:"bank_account_number" gorm:"column:bank_account_number"`
	BankAccountName   string `json:"bank_account_name" gorm:"column:bank_account_name"`
	BankName          string `json:"bank_name" gorm:"column:bank_name"`

	// Fulfillment tracking fields
	FulfillmentMode         string     `json:"fulfillment_mode" gorm:"column:fulfillment_mode;default:AUTO;check:fulfillment_mode IN ('AUTO','MANUAL')"`
	FulfillmentAttempts     int        `json:"fulfillment_attempts" gorm:"column:fulfillment_attempts;default:0"`
	LastFulfillmentAttempt  *time.Time `json:"last_fulfillment_attempt" gorm:"column:last_fulfillment_attempt"`
	FulfillmentError        string     `json:"fulfillment_error" gorm:"column:fulfillment_error"`
	CanRetry                bool       `json:"can_retry" gorm:"column:can_retry;default:true"`
	ProvisionStartedAt      *time.Time `json:"provision_started_at" gorm:"column:provision_started_at"`
	ProvisionCompletedAt    *time.Time `json:"provision_completed_at" gorm:"column:provision_completed_at"` 

	// Timestamps
	CreatedAt time.Time      `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	ExpiresAt *time.Time     `json:"expires_at" gorm:"column:expires_at"` // Default NOW() + 30 days

	// Associations
	User        *User        `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Transaction *Transaction `json:"transaction,omitempty" gorm:"foreignKey:TransactionID"`
	Prize       *WheelPrize  `json:"prize,omitempty" gorm:"foreignKey:PrizeID"`
}

// TableName specifies the table name for SpinResult
func (SpinResult) TableName() string {
	return "spin_results"
}

// BeforeCreate hook
func (s *SpinResult) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	// Set expiry to 30 days from now if not set
	if s.ExpiresAt == nil {
		expiresAt := time.Now().AddDate(0, 0, 30)
		s.ExpiresAt = &expiresAt
	}
	return nil
}

// IsPending checks if spin result is pending claim
func (s *SpinResult) IsPending() bool {
	return s.ClaimStatus == "PENDING"
}

// IsClaimed checks if spin result is claimed
func (s *SpinResult) IsClaimed() bool {
	return s.ClaimStatus == "CLAIMED"
}

// IsExpired checks if spin result is expired
func (s *SpinResult) IsExpired() bool {
	return s.ClaimStatus == "EXPIRED" || (s.ExpiresAt != nil && time.Now().After(*s.ExpiresAt))
}
