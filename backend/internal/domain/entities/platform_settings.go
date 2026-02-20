package entities

import (
	"time"

)

// PlatformSettings represents the platform_settings table
type PlatformSettings struct {
	SettingValue string     `json:"setting_value" gorm:"column:setting_value;not null" validate:"required"`
	Description  string     `json:"description" gorm:"column:description"`
	IsPublic     *bool      `json:"is_public" gorm:"column:is_public"`
	CreatedAt    *time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt    *time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// TableName specifies the table name for PlatformSettings
func (PlatformSettings) TableName() string {
	return "platform_settings"
}
