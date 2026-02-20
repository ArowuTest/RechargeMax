package entities

import (
	"time"

	"gorm.io/datatypes"
)

// UserNotifications represents the user_notifications table
type UserNotifications struct {
	Title               string         `json:"title" gorm:"column:title;not null" validate:"required"`
	Body                string         `json:"body" gorm:"column:body;not null" validate:"required"`
	NotificationType    string         `json:"notification_type" gorm:"column:notification_type;not null" validate:"required"`
	ReferenceId         string         `json:"reference_id" gorm:"column:reference_id"`
	ReferenceType       string         `json:"reference_type" gorm:"column:reference_type"`
	Channels            datatypes.JSON `json:"channels" gorm:"column:channels"`
	IsRead              *bool          `json:"is_read" gorm:"column:is_read"`
	ReadAt              *time.Time     `json:"read_at" gorm:"column:read_at"`
	DeliveryStatus      datatypes.JSON `json:"delivery_status" gorm:"column:delivery_status"`
	DeliveryAttempts    *int           `json:"delivery_attempts" gorm:"column:delivery_attempts"`
	LastDeliveryAttempt *time.Time     `json:"last_delivery_attempt" gorm:"column:last_delivery_attempt"`
	ScheduledFor        *time.Time     `json:"scheduled_for" gorm:"column:scheduled_for"`
	ExpiresAt           *time.Time     `json:"expires_at" gorm:"column:expires_at"`
	CreatedAt           *time.Time     `json:"created_at" gorm:"column:created_at"`
	UpdatedAt           *time.Time     `json:"updated_at" gorm:"column:updated_at"`
}

// TableName specifies the table name for UserNotifications
func (UserNotifications) TableName() string {
	return "user_notifications"
}
