package entities

import (
	"time"

	"github.com/google/uuid"
)

// Winner represents the winners table
type Winner struct {
	ID                   uuid.UUID      `json:"id" gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	DrawID               *uuid.UUID     `json:"draw_id" gorm:"column:draw_id;index"`
	MSISDN               string         `json:"msisdn" gorm:"column:msisdn;not null;index" validate:"required"`
	FirstName            *string        `json:"first_name" gorm:"column:first_name"`
	LastName             *string        `json:"last_name" gorm:"column:last_name"`
	Position             int            `json:"position" gorm:"column:position;not null;index" validate:"required,gt=0"`
	
	// Prize details
	PrizeType            string         `json:"prize_type" gorm:"column:prize_type;not null;index" validate:"required,oneof=cash data airtime goods"`
	PrizeDescription     string         `json:"prize_description" gorm:"column:prize_description;not null" validate:"required"`
	PrizeAmount          *int64         `json:"prize_amount" gorm:"column:prize_amount"` // For cash prizes (in kobo)
	
	// For data/airtime prizes
	DataPackage          *string        `json:"data_package" gorm:"column:data_package"`
	AirtimeAmount        *int64         `json:"airtime_amount" gorm:"column:airtime_amount"`
	Network              *string        `json:"network" gorm:"column:network"`
	
	// Auto-provisioning
	AutoProvision        bool           `json:"auto_provision" gorm:"column:auto_provision;default:false;not null"`
	ProvisionStatus      *string        `json:"provision_status" gorm:"column:provision_status;index"` // 'pending', 'provisioned', 'failed'
	ProvisionReference   *string        `json:"provision_reference" gorm:"column:provision_reference"`
	ProvisionedAt        *time.Time     `json:"provisioned_at" gorm:"column:provisioned_at"`
	ProvisionError       *string        `json:"provision_error" gorm:"column:provision_error"`
	
	// Claim management
	ClaimStatus          string         `json:"claim_status" gorm:"column:claim_status;default:PENDING;not null;index" validate:"required,oneof=PENDING CLAIMED EXPIRED REJECTED APPROVED PENDING_ADMIN_REVIEW"`
	ClaimDeadline        *time.Time     `json:"claim_deadline" gorm:"column:claim_deadline;index"`
	ClaimedAt            *time.Time     `json:"claimed_at" gorm:"column:claimed_at"`
	
	// Payout (for cash)
	PayoutStatus         string         `json:"payout_status" gorm:"column:payout_status;default:pending;not null;index" validate:"required,oneof=pending processing completed failed"`
	PayoutMethod         *string        `json:"payout_method" gorm:"column:payout_method"`
	BankCode             *string        `json:"bank_code" gorm:"column:bank_code"`
	BankName             *string        `json:"bank_name" gorm:"column:bank_name"`
	AccountNumber        *string        `json:"account_number" gorm:"column:account_number"`
	AccountName          *string        `json:"account_name" gorm:"column:account_name"`
	PayoutReference      *string        `json:"payout_reference" gorm:"column:payout_reference"`
	PayoutError          *string        `json:"payout_error" gorm:"column:payout_error"`
	
	// Goods fulfillment
	ShippingAddress      *string        `json:"shipping_address" gorm:"column:shipping_address"`
	ShippingPhone        *string        `json:"shipping_phone" gorm:"column:shipping_phone"`
	ShippingStatus       *string        `json:"shipping_status" gorm:"column:shipping_status"` // 'pending', 'shipped', 'delivered'
	TrackingNumber       *string        `json:"tracking_number" gorm:"column:tracking_number"`
	ShippedAt            *time.Time     `json:"shipped_at" gorm:"column:shipped_at"`
	DeliveredAt          *time.Time     `json:"delivered_at" gorm:"column:delivered_at"`
	
	// Notifications
	NotificationSent     bool           `json:"notification_sent" gorm:"column:notification_sent;default:false;not null"`
	NotificationSentAt   *time.Time     `json:"notification_sent_at" gorm:"column:notification_sent_at"`
	NotificationChannels *string        `json:"notification_channels" gorm:"column:notification_channels"` // JSON array
	
	// Admin notes
	Notes                *string        `json:"notes" gorm:"column:notes"`
	
	// Timestamps
	CreatedAt            time.Time      `json:"created_at" gorm:"column:created_at;autoCreateTime;index"`
	UpdatedAt            time.Time      `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`

	// Associations
	Draw *Draw `json:"draw,omitempty" gorm:"foreignKey:DrawID"`
}

// TableName specifies the table name for Winner
func (Winner) TableName() string {
	return "draw_winners"
}
