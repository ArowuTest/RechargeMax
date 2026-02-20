package entities

import (
	"time"

	"github.com/google/uuid"
)

// Wallet represents the wallets table
type Wallet struct {
	ID               uuid.UUID       `json:"id" gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	MSISDN           string          `json:"msisdn" gorm:"column:msisdn;uniqueIndex;not null" validate:"required"`
	Balance          int64           `json:"balance" gorm:"column:balance;default:0;not null"`                 // Available balance in kobo
	PendingBalance   int64           `json:"pending_balance" gorm:"column:pending_balance;default:0;not null"` // Pending earnings in kobo
	TotalEarned      int64           `json:"total_earned" gorm:"column:total_earned;default:0;not null"`       // Lifetime earnings in kobo
	TotalWithdrawn   int64           `json:"total_withdrawn" gorm:"column:total_withdrawn;default:0;not null"` // Lifetime withdrawals in kobo
	MinPayoutAmount  int64           `json:"min_payout_amount" gorm:"column:min_payout_amount;default:100000;not null"` // ₦1000 minimum
	IsActive         bool            `json:"is_active" gorm:"column:is_active;default:true;not null"`
	IsSuspended      bool            `json:"is_suspended" gorm:"column:is_suspended;default:false;not null"`
	SuspensionReason *string         `json:"suspension_reason" gorm:"column:suspension_reason"`
	LastTransactionAt *time.Time     `json:"last_transaction_at" gorm:"column:last_transaction_at"`
	CreatedAt        time.Time       `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time       `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName specifies the table name for Wallet
func (Wallet) TableName() string {
	return "wallets"
}
