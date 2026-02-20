package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// OtpVerifications represents the otp_verifications table
type OtpVerifications struct {
	Msisdn            string         `json:"msisdn" gorm:"column:msisdn;not null" validate:"required"`
	OtpCodeHash       string         `json:"otp_code_hash" gorm:"column:otp_code_hash;not null" validate:"required"`
	Purpose           string         `json:"purpose" gorm:"column:purpose"`
	UserId            *uuid.UUID     `json:"user_id" gorm:"column:user_id"`
	IsVerified        *bool          `json:"is_verified" gorm:"column:is_verified"`
	IsExpired         *bool          `json:"is_expired" gorm:"column:is_expired"`
	IsRevoked         *bool          `json:"is_revoked" gorm:"column:is_revoked"`
	Attempts          *int           `json:"attempts" gorm:"column:attempts"`
	MaxAttempts       *int           `json:"max_attempts" gorm:"column:max_attempts"`
	LastAttemptAt     *time.Time     `json:"last_attempt_at" gorm:"column:last_attempt_at"`
	RequestIp         string         `json:"request_ip" gorm:"column:request_ip"`
	RequestUserAgent  string         `json:"request_user_agent" gorm:"column:request_user_agent"`
	DeviceFingerprint string         `json:"device_fingerprint" gorm:"column:device_fingerprint"`
	VerifiedAt        *time.Time     `json:"verified_at" gorm:"column:verified_at"`
	VerifiedIp        string         `json:"verified_ip" gorm:"column:verified_ip"`
	VerifiedUserAgent string         `json:"verified_user_agent" gorm:"column:verified_user_agent"`
	CreatedAt         *time.Time     `json:"created_at" gorm:"column:created_at"`
	ExpiresAt         time.Time      `json:"expires_at" gorm:"column:expires_at;not null" validate:"required"`
	RevokedAt         *time.Time     `json:"revoked_at" gorm:"column:revoked_at"`
	Metadata          datatypes.JSON `json:"metadata" gorm:"column:metadata"`
}

// TableName specifies the table name for OtpVerifications
func (OtpVerifications) TableName() string {
	return "otp_verifications"
}
