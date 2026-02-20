package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Transactions represents the transactions table
type Transactions struct {
	ID              uuid.UUID  `json:"id" gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	TransactionCode string     `json:"transaction_code" gorm:"column:transaction_code;uniqueIndex;size:30"`
	UserID          *uuid.UUID `json:"user_id" gorm:"column:user_id;index"`

	// Transaction details
	Msisdn          string     `json:"msisdn" gorm:"column:msisdn;not null;index" validate:"required"`
	NetworkProvider string     `json:"network_provider" gorm:"column:network_provider;not null" validate:"required"`
	RechargeType    string     `json:"recharge_type" gorm:"column:recharge_type;not null;check:recharge_type IN ('AIRTIME','DATA')" validate:"required,oneof=AIRTIME DATA"`
	Amount          int64      `json:"amount" gorm:"column:amount;type:bigint;not null" validate:"required,gt=0"` // Amount in kobo (1 Naira = 100 kobo)
	DataPlanID      *uuid.UUID `json:"data_plan_id" gorm:"column:data_plan_id"`

	// Payment details
	PaymentMethod    string `json:"payment_method" gorm:"column:payment_method;not null;check:payment_method IN ('CARD','BANK_TRANSFER','USSD','WALLET')" validate:"required,oneof=CARD BANK_TRANSFER USSD WALLET"`
	PaymentReference string `json:"payment_reference" gorm:"column:payment_reference;uniqueIndex"`
	PaymentGateway   string `json:"payment_gateway" gorm:"column:payment_gateway"`

	// Transaction status
	Status            string         `json:"status" gorm:"column:status;default:PENDING;check:status IN ('PENDING','PROCESSING','SUCCESS','FAILED','CANCELLED')" validate:"oneof=PENDING PROCESSING SUCCESS FAILED CANCELLED"`
	ProviderReference string         `json:"provider_reference" gorm:"column:provider_reference"`
	ProviderResponse  datatypes.JSON `json:"provider_response" gorm:"column:provider_response;type:jsonb"`
	FailureReason     string         `json:"failure_reason" gorm:"column:failure_reason"`

	// Rewards and gamification
	PointsEarned int  `json:"points_earned" gorm:"column:points_earned;default:0"`
	DrawEntries  int  `json:"draw_entries" gorm:"column:draw_entries;default:0"`
	SpinEligible bool `json:"spin_eligible" gorm:"column:spin_eligible;default:false"`

	// Customer information (for guest transactions)
	CustomerEmail string `json:"customer_email" gorm:"column:customer_email" validate:"omitempty,email"`
	CustomerName  string `json:"customer_name" gorm:"column:customer_name"`

	// Metadata
	IpAddress     *string `json:"ip_address" gorm:"column:ip_address;type:inet"`
	UserAgent     string `json:"user_agent" gorm:"column:user_agent"`
	AffiliateCode string `json:"affiliate_code" gorm:"column:affiliate_code"`

	// Timestamps
	CreatedAt   time.Time      `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	CompletedAt *time.Time     `json:"completed_at" gorm:"column:completed_at"`
	// DeletedAt is not used in this table (no soft delete)

	// Associations
	User     *Users     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	DataPlan *DataPlans `json:"data_plan,omitempty" gorm:"foreignKey:DataPlanID"`
}

// TableName specifies the table name for Transactions
func (Transactions) TableName() string {
	return "transactions"
}

// BeforeCreate hook
func (t *Transactions) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// IsPending checks if transaction is pending
func (t *Transactions) IsPending() bool {
	return t.Status == "PENDING"
}

// IsSuccess checks if transaction is successful
func (t *Transactions) IsSuccess() bool {
	return t.Status == "SUCCESS"
}

// IsFailed checks if transaction failed
func (t *Transactions) IsFailed() bool {
	return t.Status == "FAILED"
}
