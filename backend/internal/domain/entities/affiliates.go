package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Affiliate represents the affiliates table
type Affiliate struct {
	ID     uuid.UUID  `json:"id" gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserID *uuid.UUID `json:"user_id" gorm:"column:user_id;index"`

	// Affiliate details
	AffiliateCode string `json:"affiliate_code" gorm:"column:affiliate_code;uniqueIndex;not null" validate:"required"`
	ReferralCode  string `json:"referral_code" gorm:"column:referral_code;uniqueIndex;size:20"`
	Status        string `json:"status" gorm:"column:status;default:PENDING;check:status IN ('PENDING','APPROVED','SUSPENDED','REJECTED')"`

	// Affiliate tier and commission
	Tier           string  `json:"tier" gorm:"column:tier;default:BRONZE;check:tier IN ('BRONZE','SILVER','GOLD','PLATINUM')"`
	CommissionRate float64 `json:"commission_rate" gorm:"column:commission_rate;type:decimal(5,2);default:5.00"`

	// Affiliate statistics
	TotalReferrals  int     `json:"total_referrals" gorm:"column:total_referrals;default:0"`
	ActiveReferrals int     `json:"active_referrals" gorm:"column:active_referrals;default:0"`
	TotalCommission float64 `json:"total_commission" gorm:"column:total_commission;type:decimal(10,2);default:0"`

	// Affiliate profile
	BusinessName       string         `json:"business_name" gorm:"column:business_name"`
	WebsiteUrl         string         `json:"website_url" gorm:"column:website_url"`
	SocialMediaHandles datatypes.JSON `json:"social_media_handles" gorm:"column:social_media_handles;type:jsonb"`

	// Bank details for payouts
	BankName      string `json:"bank_name" gorm:"column:bank_name"`
	AccountNumber string `json:"account_number" gorm:"column:account_number"`
	AccountName   string `json:"account_name" gorm:"column:account_name"`

	// Timestamps
	CreatedAt   time.Time      `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	ApprovedAt  *time.Time     `json:"approved_at" gorm:"column:approved_at"`

	// Associations
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for Affiliate
func (Affiliate) TableName() string {
	return "affiliates"
}

// BeforeCreate hook
func (a *Affiliate) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// IsApproved checks if affiliate is approved
func (a *Affiliate) IsApproved() bool {
	return a.Status == "APPROVED"
}

// IsSuspended checks if affiliate is suspended
func (a *Affiliate) IsSuspended() bool {
	return a.Status == "SUSPENDED"
}

// IsPending checks if affiliate is pending approval
func (a *Affiliate) IsPending() bool {
	return a.Status == "PENDING"
}
