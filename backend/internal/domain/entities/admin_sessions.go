package entities

import (
	"time"

)

// AdminSession represents the admin_sessions table
type AdminSession struct {
	SessionToken   string     `json:"session_token" gorm:"column:session_token;uniqueIndex;not null" validate:"required"`
	IpAddress      string     `json:"ip_address" gorm:"column:ip_address"`
	UserAgent      string     `json:"user_agent" gorm:"column:user_agent"`
	IsActive       *bool      `json:"is_active" gorm:"column:is_active"`
	ExpiresAt      time.Time  `json:"expires_at" gorm:"column:expires_at;not null" validate:"required"`
	CreatedAt      *time.Time `json:"created_at" gorm:"column:created_at"`
	LastAccessedAt *time.Time `json:"last_accessed_at" gorm:"column:last_accessed_at"`
}

// TableName specifies the table name for AdminSession
func (AdminSession) TableName() string {
	return "admin_sessions"
}
