package entities

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// WebhookLogs represents the webhook_logs table
type WebhookLogs struct {
	Id                             uuid.UUID                 `json:"id" gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	Source                         string                    `json:"source" gorm:"column:source;not null"`
	EventType                      string                    `json:"event_type" gorm:"column:event_type;not null"`
	Payload                        datatypes.JSON            `json:"payload" gorm:"column:payload;type:jsonb;not null"`
	Headers                        datatypes.JSON            `json:"headers" gorm:"column:headers;type:jsonb"`
	Signature                      string                    `json:"signature" gorm:"column:signature"`
	IsVerified                     *bool                     `json:"is_verified" gorm:"column:is_verified;default:false"`
	VerificationMethod             string                    `json:"verification_method" gorm:"column:verification_method"`
	VerificationError              string                    `json:"verification_error" gorm:"column:verification_error"`
	IsProcessed                    *bool                     `json:"is_processed" gorm:"column:is_processed;default:false"`
	ProcessingError                string                    `json:"processing_error" gorm:"column:processing_error"`
	ProcessingAttempts             *int                      `json:"processing_attempts" gorm:"column:processing_attempts;default:0"`
	MaxProcessingAttempts          *int                      `json:"max_processing_attempts" gorm:"column:max_processing_attempts;default:3"`
	TransactionReference           string                    `json:"transaction_reference" gorm:"column:transaction_reference"`
	RelatedTransactionId           *uuid.UUID                `json:"related_transaction_id" gorm:"column:related_transaction_id"`
	ReceivedAt                     time.Time                 `json:"received_at" gorm:"column:received_at;autoCreateTime"`
	VerifiedAt                     *time.Time                `json:"verified_at" gorm:"column:verified_at"`
	ProcessedAt                    *time.Time                `json:"processed_at" gorm:"column:processed_at"`
	NextRetryAt                    *time.Time                `json:"next_retry_at" gorm:"column:next_retry_at"`
	IpAddress                      string                    `json:"ip_address" gorm:"column:ip_address"`
	UserAgent                      string                    `json:"user_agent" gorm:"column:user_agent"`
	Metadata                       datatypes.JSON            `json:"metadata" gorm:"column:metadata;type:jsonb"`
}

// TableName specifies the table name for WebhookLogs
func (WebhookLogs) TableName() string {
	return "webhook_logs"
}
