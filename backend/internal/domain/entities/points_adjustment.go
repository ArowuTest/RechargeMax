package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PointsAdjustment represents a manual points adjustment by admin
type PointsAdjustment struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	User        *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Points      int        `gorm:"not null" json:"points"` // Positive for add, negative for deduct
	Reason      string     `gorm:"type:varchar(100);not null" json:"reason"`
	Description string     `gorm:"type:text" json:"description"`
	AdminID     uuid.UUID  `gorm:"type:uuid;not null;column:admin_id" json:"admin_id"`
	CreatedBy   uuid.UUID  `gorm:"type:uuid;not null;column:created_by" json:"created_by"`
	AdminUser   *User      `gorm:"foreignKey:AdminID" json:"admin_user,omitempty"`
	CreatedAt   time.Time  `gorm:"not null;index" json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// BeforeCreate auto-populates created_by from admin_id so both NOT NULL
// columns are satisfied (schema has created_by; entity uses admin_id).
func (p *PointsAdjustment) BeforeCreate(_ *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	if p.CreatedBy == uuid.Nil {
		p.CreatedBy = p.AdminID
	}
	return nil
}

// TableName specifies the table name for PointsAdjustment
func (PointsAdjustment) TableName() string {
	return "points_adjustments"
}
