package entities

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// WalletTransactions represents the wallet_transactions table
type WalletTransactions struct {
	Id                             uuid.UUID                 `json:"id" gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserId                         uuid.UUID                 `json:"user_id" gorm:"column:user_id;not null"`
	TransactionType                string                    `json:"transaction_type" gorm:"column:transaction_type;not null"`
	Amount                         int64                     `json:"amount" gorm:"column:amount;type:bigint;not null"` // Amount in kobo
	BalanceBefore                  int64                     `json:"balance_before" gorm:"column:balance_before;type:bigint;not null"` // Balance in kobo
	BalanceAfter                   int64                     `json:"balance_after" gorm:"column:balance_after;type:bigint;not null"` // Balance in kobo
	Source                         string                    `json:"source" gorm:"column:source;not null"`
	Reference                      string                    `json:"reference" gorm:"column:reference;uniqueIndex;not null"`
	RelatedTransactionId           *uuid.UUID                `json:"related_transaction_id" gorm:"column:related_transaction_id"`
	Description                    string                    `json:"description" gorm:"column:description"`
	Status                         string                    `json:"status" gorm:"column:status;default:'COMPLETED'"`
	Metadata                       datatypes.JSON            `json:"metadata" gorm:"column:metadata;type:jsonb"`
	AdminId                        *uuid.UUID                `json:"admin_id" gorm:"column:admin_id"`
	CreatedAt                      time.Time                 `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	CompletedAt                    *time.Time                `json:"completed_at" gorm:"column:completed_at"`
	ReversedAt                     *time.Time                `json:"reversed_at" gorm:"column:reversed_at"`
}

// TableName specifies the table name for WalletTransactions
func (WalletTransactions) TableName() string {
	return "wallet_transactions"
}
