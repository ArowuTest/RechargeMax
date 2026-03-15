package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Notification represents the notifications table
type Notification struct {
	ID            uuid.UUID      `json:"id" gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	MSISDN        string         `json:"msisdn" gorm:"column:msisdn;not null;index" validate:"required"`
	Type          string         `json:"type" gorm:"column:type;not null;index" validate:"required,oneof=draw_winner prize_claimed payout_completed spin_win commission_earned withdrawal_processed system announcement"`
	Title         string         `json:"title" gorm:"column:title;not null" validate:"required,max=200"`
	Message       string         `json:"message" gorm:"column:message;not null" validate:"required"`
	Icon          *string        `json:"icon" gorm:"column:icon" validate:"omitempty,oneof=trophy money gift info warning success error"`
	Priority      string         `json:"priority" gorm:"column:priority;default:normal;not null" validate:"required,oneof=high normal low"`
	ActionURL     *string        `json:"action_url" gorm:"column:action_url"`
	ActionLabel   *string        `json:"action_label" gorm:"column:action_label"`
	IsRead        bool           `json:"is_read" gorm:"column:is_read;default:false;not null;index"`
	ReadAt        *time.Time     `json:"read_at" gorm:"column:read_at"`
	ReferenceType *string        `json:"reference_type" gorm:"column:reference_type;index"`
	ReferenceID   *string        `json:"reference_id" gorm:"column:reference_id;index"`
	Metadata      datatypes.JSON `json:"metadata" gorm:"column:metadata;type:jsonb"`
	CreatedAt     time.Time      `json:"created_at" gorm:"column:created_at;autoCreateTime;index"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName specifies the table name for Notification
func (Notification) TableName() string {
	return "user_notifications"
}
