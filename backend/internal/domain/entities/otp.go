package entities

import (
	"time"

	"github.com/google/uuid"
)

// OTP represents an OTP (One-Time Password) record
type OTP struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	MSISDN    string     `gorm:"type:varchar(15);not null;index" json:"msisdn"`
	Code      string     `gorm:"type:varchar(10);not null" json:"code"`
	Purpose   string     `gorm:"type:varchar(50);default:'login'" json:"purpose"` // login, verification, password_reset
	ExpiresAt time.Time  `gorm:"not null;index" json:"expires_at"`
	IsUsed    bool       `gorm:"default:false;index" json:"is_used"`
	UsedAt    *time.Time `gorm:"index" json:"used_at,omitempty"`
	CreatedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	FailedAttempts int       `gorm:"default:0" json:"failed_attempts"` // SEC-008: brute-force guard
	UpdatedAt      time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName specifies the table name for OTP
func (OTP) TableName() string {
	return "otps"
}

// IsExpired checks if the OTP has expired
func (o *OTP) IsExpired() bool {
	return time.Now().After(o.ExpiresAt)
}

// IsValid checks if the OTP is valid (not used and not expired)
func (o *OTP) IsValid() bool {
	return !o.IsUsed && !o.IsExpired()
}
