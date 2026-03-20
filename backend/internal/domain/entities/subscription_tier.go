package entities

import (
	"time"

	"github.com/google/uuid"
)

// SubscriptionTier represents a configurable subscription tier
type SubscriptionTier struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name        string     `json:"name" gorm:"not null" validate:"required"`
	Description string     `json:"description"`
	Entries     int        `json:"entries" gorm:"not null" validate:"required,min=1"` // Number of draw entries
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	SortOrder   int        `json:"sort_order" gorm:"default:0"` // For display ordering
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name
func (SubscriptionTier) TableName() string {
	return "subscription_tiers"
}

// SubscriptionPricing represents the global pricing configuration
type SubscriptionPricing struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	PricePerEntry int64      `json:"price_per_entry" gorm:"not null"` // Price in kobo (₦20 = 2000 kobo)
	Currency      string     `json:"currency" gorm:"default:'NGN'"`
	IsActive      bool       `json:"is_active" gorm:"default:true"`
	EffectiveFrom time.Time  `json:"effective_from" gorm:"not null"`
	EffectiveTo   *time.Time `json:"effective_to,omitempty"`
	CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name
func (SubscriptionPricing) TableName() string {
	return "subscription_pricing"
}


// SubscriptionBilling is defined in subscription_billing.go
