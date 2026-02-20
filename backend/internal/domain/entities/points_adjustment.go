package entities

import (
	"time"

	"github.com/google/uuid"
)

// PointsAdjustment represents a manual points adjustment by admin
type PointsAdjustment struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	User        *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Points      int        `gorm:"not null" json:"points"` // Positive for add, negative for deduct
	Reason      string     `gorm:"type:varchar(100);not null" json:"reason"`
	Description string     `gorm:"type:text" json:"description"`
	AdminID     uuid.UUID  `gorm:"type:uuid;not null" json:"admin_id"`
	AdminUser   *User      `gorm:"foreignKey:AdminID" json:"admin_user,omitempty"`
	CreatedAt   time.Time  `gorm:"not null;index" json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TableName specifies the table name for PointsAdjustment
func (PointsAdjustment) TableName() string {
	return "points_adjustments"
}
