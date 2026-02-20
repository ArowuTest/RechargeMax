package entities

import (
	"time"

	"github.com/google/uuid"
)

// Device represents the devices table
type Device struct {
	ID                     uuid.UUID      `json:"id" gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	MSISDN                 string         `json:"msisdn" gorm:"column:msisdn;not null;index" validate:"required"`
	DeviceID               string         `json:"device_id" gorm:"column:device_id;uniqueIndex;not null" validate:"required"`
	FCMToken               *string        `json:"fcm_token" gorm:"column:fcm_token;index"`
	Platform               string         `json:"platform" gorm:"column:platform;not null" validate:"required,oneof=ios android web"`
	AppVersion             *string        `json:"app_version" gorm:"column:app_version"`
	DeviceModel            *string        `json:"device_model" gorm:"column:device_model"`
	OSVersion              *string        `json:"os_version" gorm:"column:os_version"`
	IsActive               bool           `json:"is_active" gorm:"column:is_active;default:true;not null;index"`
	LastActive             time.Time      `json:"last_active" gorm:"column:last_active;default:CURRENT_TIMESTAMP;not null;index"`
	LastNotificationSentAt *time.Time     `json:"last_notification_sent_at" gorm:"column:last_notification_sent_at"`
	NotificationCount      int            `json:"notification_count" gorm:"column:notification_count;default:0;not null"`
	CreatedAt              time.Time      `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt              time.Time      `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName specifies the table name for Device
func (Device) TableName() string {
	return "devices"
}
