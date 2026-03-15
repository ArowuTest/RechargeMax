package entities

import (
	"time"

	"gorm.io/datatypes"
)

// AdminActivityLog represents the admin_activity_logs table
type AdminActivityLog struct {
	Action         string         `json:"action" gorm:"column:action;not null" validate:"required"`
	Resource       string         `json:"resource" gorm:"column:resource"`
	ResourceId     string         `json:"resource_id" gorm:"column:resource_id"`
	Method         string         `json:"method" gorm:"column:method"`
	Endpoint       string         `json:"endpoint" gorm:"column:endpoint"`
	RequestData    datatypes.JSON `json:"request_data" gorm:"column:request_data"`
	ResponseStatus *int           `json:"response_status" gorm:"column:response_status"`
	ResponseData   datatypes.JSON `json:"response_data" gorm:"column:response_data"`
	IpAddress      string         `json:"ip_address" gorm:"column:ip_address"`
	UserAgent      string         `json:"user_agent" gorm:"column:user_agent"`
	DurationMs     *int           `json:"duration_ms" gorm:"column:duration_ms"`
	IsSuspicious   *bool          `json:"is_suspicious" gorm:"column:is_suspicious"`
	CreatedAt      *time.Time     `json:"created_at" gorm:"column:created_at"`
}

// TableName specifies the table name for AdminActivityLog
func (AdminActivityLog) TableName() string {
	return "admin_activity_logs"
}
