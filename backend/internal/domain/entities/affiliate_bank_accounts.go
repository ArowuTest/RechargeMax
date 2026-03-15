package entities

import (
	"time"

)

// AffiliateBankAccount represents the affiliate_bank_accounts table
type AffiliateBankAccount struct {
	BankName      string     `json:"bank_name" gorm:"column:bank_name;not null" validate:"required"`
	AccountNumber string     `json:"account_number" gorm:"column:account_number;not null" validate:"required"`
	AccountName   string     `json:"account_name" gorm:"column:account_name;not null" validate:"required"`
	IsVerified    *bool      `json:"is_verified" gorm:"column:is_verified"`
	IsPrimary     *bool      `json:"is_primary" gorm:"column:is_primary"`
	VerifiedAt    *time.Time `json:"verified_at" gorm:"column:verified_at"`
	IsActive      *bool      `json:"is_active" gorm:"column:is_active"`
	CreatedAt     *time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt     *time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// TableName specifies the table name for AffiliateBankAccount
func (AffiliateBankAccount) TableName() string {
	return "affiliate_bank_accounts"
}
