package entities

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// ApiLogs represents the api_logs table
type ApiLogs struct {
	Id                             uuid.UUID                 `json:"id" gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	ServiceName                    string                    `json:"service_name" gorm:"column:service_name;not null"`
	Endpoint                       string                    `json:"endpoint" gorm:"column:endpoint;not null"`
	Method                         string                    `json:"method" gorm:"column:method;not null"`
	RequestUrl                     string                    `json:"request_url" gorm:"column:request_url"`
	RequestHeaders                 datatypes.JSON            `json:"request_headers" gorm:"column:request_headers;type:jsonb"`
	RequestPayload                 datatypes.JSON            `json:"request_payload" gorm:"column:request_payload;type:jsonb"`
	ResponseStatusCode             *int                      `json:"response_status_code" gorm:"column:response_status_code"`
	ResponseHeaders                datatypes.JSON            `json:"response_headers" gorm:"column:response_headers;type:jsonb"`
	ResponsePayload                datatypes.JSON            `json:"response_payload" gorm:"column:response_payload;type:jsonb"`
	ResponseTimeMs                 *int                      `json:"response_time_ms" gorm:"column:response_time_ms"`
	IsError                        *bool                     `json:"is_error" gorm:"column:is_error;default:false"`
	ErrorMessage                   string                    `json:"error_message" gorm:"column:error_message"`
	ErrorCode                      string                    `json:"error_code" gorm:"column:error_code"`
	UserId                         *uuid.UUID                `json:"user_id" gorm:"column:user_id"`
	TransactionReference           string                    `json:"transaction_reference" gorm:"column:transaction_reference"`
	IpAddress                      string                    `json:"ip_address" gorm:"column:ip_address"`
	CreatedAt                      time.Time                 `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	Metadata                       datatypes.JSON            `json:"metadata" gorm:"column:metadata;type:jsonb"`
}

// TableName specifies the table name for ApiLogs
func (ApiLogs) TableName() string {
	return "api_logs"
}
