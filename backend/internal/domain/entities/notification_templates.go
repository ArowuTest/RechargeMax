package entities

import (
	"time"

	"gorm.io/datatypes"
)

// NotificationTemplates represents the notification_templates table
type NotificationTemplates struct {
	TemplateKey          string         `json:"template_key" gorm:"column:template_key;uniqueIndex;not null" validate:"required"`
	TemplateName         string         `json:"template_name" gorm:"column:template_name;not null" validate:"required"`
	Description          string         `json:"description" gorm:"column:description"`
	TitleTemplate        string         `json:"title_template" gorm:"column:title_template;not null" validate:"required"`
	BodyTemplate         string         `json:"body_template" gorm:"column:body_template;not null" validate:"required"`
	EmailSubjectTemplate string         `json:"email_subject_template" gorm:"column:email_subject_template" validate:"email"`
	EmailBodyTemplate    string         `json:"email_body_template" gorm:"column:email_body_template" validate:"email"`
	SmsTemplate          string         `json:"sms_template" gorm:"column:sms_template"`
	Variables            datatypes.JSON `json:"variables" gorm:"column:variables"`
	SupportsPush         *bool          `json:"supports_push" gorm:"column:supports_push"`
	SupportsEmail        *bool          `json:"supports_email" gorm:"column:supports_email" validate:"email"`
	SupportsSms          *bool          `json:"supports_sms" gorm:"column:supports_sms"`
	SupportsInApp        *bool          `json:"supports_in_app" gorm:"column:supports_in_app"`
	IsActive             *bool          `json:"is_active" gorm:"column:is_active"`
	CreatedAt            *time.Time     `json:"created_at" gorm:"column:created_at"`
	UpdatedAt            *time.Time     `json:"updated_at" gorm:"column:updated_at"`
}

// TableName specifies the table name for NotificationTemplates
func (NotificationTemplates) TableName() string {
	return "notification_templates"
}
