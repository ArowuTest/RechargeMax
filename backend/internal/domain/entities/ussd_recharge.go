package entities

import (
	"time"

	"github.com/google/uuid"
)

// USSDRecharge represents a recharge made via USSD directly with telecom provider
type USSDRecharge struct {
	ID                 uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID             *uuid.UUID `json:"user_id" gorm:"type:uuid;index"`
	MSISDN             string     `json:"msisdn" gorm:"not null;index" validate:"required"`
	Network            string     `json:"network" gorm:"not null"` // MTN, Glo, Airtel, 9mobile
	Amount             int64      `json:"amount" gorm:"not null"` // Amount in kobo
	RechargeType       string     `json:"recharge_type" gorm:"not null"` // airtime, data
	ProductCode        string     `json:"product_code"` // Bundle code if data
	TransactionRef     string     `json:"transaction_ref" gorm:"uniqueIndex;not null"` // From telecom provider
	ProviderRef        string     `json:"provider_ref"` // Additional reference from provider
	PointsEarned       int        `json:"points_earned" gorm:"default:0"` // ₦200 = 1 point
	Status             string     `json:"status" gorm:"default:'completed'"` // completed, failed
	RechargeDate       time.Time  `json:"recharge_date" gorm:"not null;index"`
	ReceivedAt         time.Time  `json:"received_at" gorm:"not null"` // When webhook received
	ProcessedAt        *time.Time `json:"processed_at,omitempty"` // When points allocated
	WebhookPayload     string     `json:"webhook_payload" gorm:"type:text"` // Raw webhook data for debugging
	Notes              string     `json:"notes" gorm:"type:text"`
	CreatedAt          time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name
func (USSDRecharge) TableName() string {
	return "ussd_recharges"
}

// USSDWebhookLog represents a log of all USSD webhooks received
type USSDWebhookLog struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Provider       string     `json:"provider" gorm:"not null;index"` // MTN, Glo, Airtel, 9mobile
	Endpoint       string     `json:"endpoint" gorm:"not null"` // Which webhook endpoint
	Method         string     `json:"method" gorm:"not null"` // GET, POST
	Headers        string     `json:"headers" gorm:"type:text"` // Request headers
	Body           string     `json:"body" gorm:"type:text"` // Request body
	IPAddress      string     `json:"ip_address"`
	Status         string     `json:"status" gorm:"default:'received'"` // received, processed, failed
	ProcessingError string    `json:"processing_error" gorm:"type:text"`
	USSDRechargeID *uuid.UUID `json:"ussd_recharge_id" gorm:"type:uuid;index"` // Link to created recharge
	ReceivedAt     time.Time  `json:"received_at" gorm:"not null;index"`
	ProcessedAt    *time.Time `json:"processed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

// TableName specifies the table name
func (USSDWebhookLog) TableName() string {
	return "ussd_webhook_logs"
}
