package entities

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// BankAccount represents the bank_accounts table
type BankAccount struct {
	Id                             uuid.UUID                 `json:"id" gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserId                         uuid.UUID                 `json:"user_id" gorm:"column:user_id;not null"`
	AccountName                    string                    `json:"account_name" gorm:"column:account_name;not null"`
	AccountNumber                  string                    `json:"account_number" gorm:"column:account_number;not null"`
	BankName                       string                    `json:"bank_name" gorm:"column:bank_name;not null"`
	BankCode                       string                    `json:"bank_code" gorm:"column:bank_code;not null"`
	IsVerified                     *bool                     `json:"is_verified" gorm:"column:is_verified;default:false"`
	IsPrimary                      *bool                     `json:"is_primary" gorm:"column:is_primary;default:false"`
	VerificationMethod             string                    `json:"verification_method" gorm:"column:verification_method"`
	VerificationData               datatypes.JSON            `json:"verification_data" gorm:"column:verification_data;type:jsonb"`
	CreatedAt                      time.Time                 `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	VerifiedAt                     *time.Time                `json:"verified_at" gorm:"column:verified_at"`
	LastUsedAt                     *time.Time                `json:"last_used_at" gorm:"column:last_used_at"`
}

// TableName specifies the table name for BankAccount
func (BankAccount) TableName() string {
	return "bank_accounts"
}
