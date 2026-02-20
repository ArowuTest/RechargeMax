package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// WalletTransaction represents the wallet_transactions table
type WalletTransaction struct {
	ID             uuid.UUID      `json:"id" gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	WalletID       uuid.UUID      `json:"wallet_id" gorm:"column:wallet_id;not null;index" validate:"required"`
	TransactionID  string         `json:"transaction_id" gorm:"column:transaction_id;uniqueIndex;not null" validate:"required"`
	Type           string         `json:"type" gorm:"column:type;not null" validate:"required,oneof=credit debit pending_credit pending_release"`
	Amount         int64          `json:"amount" gorm:"column:amount;not null" validate:"required,gt=0"` // Amount in kobo
	BalanceBefore  int64          `json:"balance_before" gorm:"column:balance_before;not null"`
	BalanceAfter   int64          `json:"balance_after" gorm:"column:balance_after;not null"`
	Description    string         `json:"description" gorm:"column:description;not null" validate:"required"`
	ReferenceType  *string        `json:"reference_type" gorm:"column:reference_type;index"` // 'referral_commission', 'payout', 'prize_win', 'adjustment'
	ReferenceID    *string        `json:"reference_id" gorm:"column:reference_id;index"`
	Status         string         `json:"status" gorm:"column:status;default:completed;not null" validate:"required,oneof=pending completed failed reversed"`
	Metadata       datatypes.JSON `json:"metadata" gorm:"column:metadata;type:jsonb"`
	ProcessedAt    time.Time      `json:"processed_at" gorm:"column:processed_at;default:CURRENT_TIMESTAMP;not null"`
	CreatedAt      time.Time      `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`

	// Associations
	Wallet *Wallet `json:"wallet,omitempty" gorm:"foreignKey:WalletID"`
}

// TableName specifies the table name for WalletTransaction
func (WalletTransaction) TableName() string {
	return "wallet_transactions"
}
