package entities

import (
	"time"

	"gorm.io/datatypes"
)

// ApplicationLog represents the application_logs table
type ApplicationLog struct {
	Message    string         `json:"message" gorm:"column:message;not null" validate:"required"`
	Context    datatypes.JSON `json:"context" gorm:"column:context"`
	IpAddress  string         `json:"ip_address" gorm:"column:ip_address"`
	UserAgent  string         `json:"user_agent" gorm:"column:user_agent"`
	RequestID  string         `json:"request_id" gorm:"column:request_id"`
	ErrorCode  string         `json:"error_code" gorm:"column:error_code"`
	StackTrace string         `json:"stack_trace" gorm:"column:stack_trace"`
	CreatedAt  *time.Time     `json:"created_at" gorm:"column:created_at"`
}

// TableName specifies the table name for ApplicationLog
func (ApplicationLog) TableName() string {
	return "application_logs"
}
