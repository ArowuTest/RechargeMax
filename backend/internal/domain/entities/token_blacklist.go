package entities

import (
	"time"

	"github.com/google/uuid"
)

// TokenBlacklist represents revoked JWT tokens
type TokenBlacklist struct {
	ID        uuid.UUID `json:"id" gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	Token     string    `json:"token" gorm:"column:token;not null;uniqueIndex" validate:"required"`
	AdminID   uuid.UUID `json:"admin_id" gorm:"column:admin_id;not null;index" validate:"required"`
	Reason    string    `json:"reason" gorm:"column:reason;not null" validate:"required"`
	ExpiresAt time.Time `json:"expires_at" gorm:"column:expires_at;not null;index" validate:"required"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime;index"`
}

// TableName specifies the table name for TokenBlacklist
func (TokenBlacklist) TableName() string {
	return "token_blacklist"
}

// IsExpired checks if the token expiry has passed
func (t *TokenBlacklist) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}
