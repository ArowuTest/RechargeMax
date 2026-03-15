package entities

import (
	"time"

	"gorm.io/datatypes"
)

// NotificationDeliveryLog represents the notification_delivery_log table
type NotificationDeliveryLog struct {
	Channel           string         `json:"channel" gorm:"column:channel;not null" validate:"required"`
	DeliveryStatus    string         `json:"delivery_status" gorm:"column:delivery_status;not null" validate:"required"`
	ProviderName      string         `json:"provider_name" gorm:"column:provider_name"`
	ProviderMessageID string         `json:"provider_message_id" gorm:"column:provider_message_id"`
	ProviderResponse  datatypes.JSON `json:"provider_response" gorm:"column:provider_response"`
	ErrorCode         string         `json:"error_code" gorm:"column:error_code"`
	ErrorMessage      string         `json:"error_message" gorm:"column:error_message"`
	RetryCount        *int           `json:"retry_count" gorm:"column:retry_count"`
	AttemptedAt       *time.Time     `json:"attempted_at" gorm:"column:attempted_at"`
	DeliveredAt       *time.Time     `json:"delivered_at" gorm:"column:delivered_at"`
}

// TableName specifies the table name for NotificationDeliveryLog
func (NotificationDeliveryLog) TableName() string {
	return "notification_delivery_log"
}
