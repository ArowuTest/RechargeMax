package entities

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// WithdrawalRequest represents the withdrawal_requests table
type WithdrawalRequest struct {
	ID                             uuid.UUID                 `json:"id" gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserID                         uuid.UUID                 `json:"user_id" gorm:"column:user_id;not null"`
	BankAccountID                  uuid.UUID                 `json:"bank_account_id" gorm:"column:bank_account_id;not null"`
	Amount                         int64                     `json:"amount" gorm:"column:amount;type:bigint;not null"` // Amount in kobo
	Fee                            int64                     `json:"fee" gorm:"column:fee;type:bigint;default:0"` // Fee in kobo
	NetAmount                      int64                     `json:"net_amount" gorm:"column:net_amount;type:bigint;not null"` // Net amount in kobo
	Status                         string                    `json:"status" gorm:"column:status;default:'PENDING'"`
	ApprovedByAdminID              *uuid.UUID                `json:"approved_by_admin_id" gorm:"column:approved_by_admin_id"`
	RejectionReason                string                    `json:"rejection_reason" gorm:"column:rejection_reason"`
	AdminNotes                     string                    `json:"admin_notes" gorm:"column:admin_notes"`
	TransactionReference           string                    `json:"transaction_reference" gorm:"column:transaction_reference;uniqueIndex"`
	BankReference                  string                    `json:"bank_reference" gorm:"column:bank_reference"`
	PaymentProvider                string                    `json:"payment_provider" gorm:"column:payment_provider"`
	ProviderResponse               datatypes.JSON            `json:"provider_response" gorm:"column:provider_response;type:jsonb"`
	WalletTransactionID            *uuid.UUID                `json:"wallet_transaction_id" gorm:"column:wallet_transaction_id"`
	RequestedAt                    time.Time                 `json:"requested_at" gorm:"column:requested_at;autoCreateTime"`
	ApprovedAt                     *time.Time                `json:"approved_at" gorm:"column:approved_at"`
	ProcessingStartedAt            *time.Time                `json:"processing_started_at" gorm:"column:processing_started_at"`
	CompletedAt                    *time.Time                `json:"completed_at" gorm:"column:completed_at"`
	RejectedAt                     *time.Time                `json:"rejected_at" gorm:"column:rejected_at"`
	RequestIp                      string                    `json:"request_ip" gorm:"column:request_ip"`
	RequestUserAgent               string                    `json:"request_user_agent" gorm:"column:request_user_agent"`
}

// TableName specifies the table name for WithdrawalRequest
func (WithdrawalRequest) TableName() string {
	return "withdrawal_requests"
}
