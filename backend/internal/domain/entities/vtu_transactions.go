package entities

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// VtuTransaction represents the vtu_transactions table
type VtuTransaction struct {
	Id                             uuid.UUID                 `json:"id" gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	TransactionReference           string                    `json:"transaction_reference" gorm:"column:transaction_reference;uniqueIndex;not null"`
	ParentTransactionId            *uuid.UUID                `json:"parent_transaction_id" gorm:"column:parent_transaction_id"`
	UserId                         *uuid.UUID                `json:"user_id" gorm:"column:user_id"`
	PhoneNumber                    string                    `json:"phone_number" gorm:"column:phone_number;not null"`
	NetworkProvider                string                    `json:"network_provider" gorm:"column:network_provider;not null"`
	RechargeType                   string                    `json:"recharge_type" gorm:"column:recharge_type;not null"`
	Amount                         int64                     `json:"amount" gorm:"column:amount;type:bigint;not null"` // Amount in kobo (1 Naira = 100 kobo)
	DataBundle                     string                    `json:"data_bundle" gorm:"column:data_bundle"`
	DataBundleCode                 string                    `json:"data_bundle_code" gorm:"column:data_bundle_code"`
	ProviderUsed                   string                    `json:"provider_used" gorm:"column:provider_used"`
	ProviderTransactionId          string                    `json:"provider_transaction_id" gorm:"column:provider_transaction_id"`
	ProviderReference              string                    `json:"provider_reference" gorm:"column:provider_reference"`
	ProviderResponse               datatypes.JSON            `json:"provider_response" gorm:"column:provider_response;type:jsonb"`
	ProviderStatus                 string                    `json:"provider_status" gorm:"column:provider_status"`
	Status                         string                    `json:"status" gorm:"column:status;default:'PENDING'"`
	RetryCount                     *int                      `json:"retry_count" gorm:"column:retry_count;default:0"`
	MaxRetries                     *int                      `json:"max_retries" gorm:"column:max_retries;default:3"`
	UserAgent                      string                    `json:"user_agent" gorm:"column:user_agent"`
	IpAddress                      string                    `json:"ip_address" gorm:"column:ip_address"`
	DeviceInfo                     datatypes.JSON            `json:"device_info" gorm:"column:device_info;type:jsonb"`
	CreatedAt                      time.Time                 `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	ProcessingStartedAt            *time.Time                `json:"processing_started_at" gorm:"column:processing_started_at"`
	CompletedAt                    *time.Time                `json:"completed_at" gorm:"column:completed_at"`
	FailedAt                       *time.Time                `json:"failed_at" gorm:"column:failed_at"`
	ErrorMessage                   string                    `json:"error_message" gorm:"column:error_message"`
	ErrorCode                      string                    `json:"error_code" gorm:"column:error_code"`
	LastErrorAt                    *time.Time                `json:"last_error_at" gorm:"column:last_error_at"`
	IsReconciled                   *bool                     `json:"is_reconciled" gorm:"column:is_reconciled;default:false"`
	ReconciledAt                   *time.Time                `json:"reconciled_at" gorm:"column:reconciled_at"`
	ReconciliationNotes            string                    `json:"reconciliation_notes" gorm:"column:reconciliation_notes"`
}

// TableName specifies the table name for VtuTransaction
func (VtuTransaction) TableName() string {
	return "vtu_transactions"
}
