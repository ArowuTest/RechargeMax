package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// NotificationDeliveryLog represents the notification_delivery_log table.
// Every SMS/push/email send attempt is logged here with outcome + provider response.
type NotificationDeliveryLog struct {
	ID                uuid.UUID      `json:"id" gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	NotificationID    *uuid.UUID     `json:"notification_id" gorm:"column:notification_id;type:uuid"`
	Channel           string         `json:"channel" gorm:"column:channel;not null" validate:"required,oneof=push email sms in_app"`
	DeliveryStatus    string         `json:"delivery_status" gorm:"column:delivery_status;not null" validate:"required,oneof=pending sent delivered failed bounced opened clicked"`
	ProviderName      string         `json:"provider_name" gorm:"column:provider_name"`
	ProviderMessageID string         `json:"provider_message_id" gorm:"column:provider_message_id"`
	ProviderResponse  datatypes.JSON `json:"provider_response" gorm:"column:provider_response;type:jsonb"`
	ErrorCode         string         `json:"error_code" gorm:"column:error_code"`
	ErrorMessage      string         `json:"error_message" gorm:"column:error_message"`
	RetryCount        *int           `json:"retry_count" gorm:"column:retry_count;default:0"`
	AttemptedAt       *time.Time     `json:"attempted_at" gorm:"column:attempted_at;autoCreateTime"`
	DeliveredAt       *time.Time     `json:"delivered_at" gorm:"column:delivered_at"`
}

func (NotificationDeliveryLog) TableName() string {
	return "notification_delivery_log"
}
