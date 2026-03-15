package entities

import (
	"time"
)

// PlatformSetting represents the platform_settings table
type PlatformSetting struct {
	SettingKey   string     `json:"setting_key" gorm:"column:setting_key;primaryKey"`
	SettingValue string     `json:"setting_value" gorm:"column:setting_value;not null" validate:"required"`
	Description  string     `json:"description" gorm:"column:description"`
	IsPublic     *bool      `json:"is_public" gorm:"column:is_public"`
	CreatedAt    *time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt    *time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// TableName specifies the table name for PlatformSetting
func (PlatformSetting) TableName() string {
	return "platform_settings"
}
