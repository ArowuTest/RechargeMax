package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Users represents the users table
type Users struct {
	ID         uuid.UUID  `json:"id" gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserCode   string     `json:"user_code" gorm:"column:user_code;uniqueIndex;size:20"`
	AuthUserID *uuid.UUID `json:"auth_user_id" gorm:"column:auth_user_id;uniqueIndex"`

	// Basic Information
	MSISDN        string `json:"msisdn" gorm:"column:msisdn;uniqueIndex;not null" validate:"required"`
	FullName      string `json:"full_name" gorm:"column:full_name"`
	Email         string `json:"email" gorm:"column:email" validate:"omitempty,email"`
	PhoneVerified bool   `json:"phone_verified" gorm:"column:phone_verified;default:false"`
	EmailVerified bool   `json:"email_verified" gorm:"column:email_verified;default:false"`

	// Profile Information
	DateOfBirth *time.Time `json:"date_of_birth" gorm:"column:date_of_birth;type:date"`
	Gender      string     `json:"gender" gorm:"column:gender"`
	State       string     `json:"state" gorm:"column:state"`
	City        string     `json:"city" gorm:"column:city"`
	Address     string     `json:"address" gorm:"column:address"`

	// Gamification and Loyalty
	TotalPoints         int     `json:"total_points" gorm:"column:total_points;default:0"`
	LoyaltyTier         string  `json:"loyalty_tier" gorm:"column:loyalty_tier;default:'BRONZE'"`
	TotalRechargeAmount int64 `json:"total_recharge_amount" gorm:"column:total_recharge_amount;type:bigint;default:0"` // Total in kobo
	TotalTransactions   int     `json:"total_transactions" gorm:"column:total_transactions;default:0"`
	LastRechargeDate    *time.Time `json:"last_recharge_date" gorm:"column:last_recharge_date"`

	// Referral System
	ReferralCode   string     `json:"referral_code" gorm:"column:referral_code;uniqueIndex"`
	ReferredBy     *uuid.UUID `json:"referred_by" gorm:"column:referred_by"`
	TotalReferrals int        `json:"total_referrals" gorm:"column:total_referrals;default:0"`

	// Account Status
	IsActive   bool   `json:"is_active" gorm:"column:is_active;default:true"`
	IsVerified bool   `json:"is_verified" gorm:"column:is_verified;default:false"`
	KYCStatus  string `json:"kyc_status" gorm:"column:kyc_status;default:'PENDING'"`

	// Timestamps
	CreatedAt   time.Time      `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	LastLoginAt *time.Time     `json:"last_login_at" gorm:"column:last_login_at"`
}

// TableName specifies the table name for Users
func (Users) TableName() string {
	return "users"
}

// BeforeCreate hook
func (u *Users) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// IsPhoneVerified checks if phone is verified
func (u *Users) IsPhoneVerified() bool {
	return u.PhoneVerified
}

// IsEmailVerified checks if email is verified
func (u *Users) IsEmailVerified() bool {
	return u.EmailVerified
}

// IsKYCVerified checks if KYC is verified
func (u *Users) IsKYCVerified() bool {
	return u.KYCStatus == "VERIFIED"
}
