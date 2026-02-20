package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// AuditLog represents the audit_logs table
type AuditLog struct {
	ID           uuid.UUID      `json:"id" gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	AdminUserID  uuid.UUID      `json:"admin_user_id" gorm:"column:admin_user_id;not null;index" validate:"required"`
	AdminEmail   *string        `json:"admin_email" gorm:"column:admin_email"`
	AdminName    *string        `json:"admin_name" gorm:"column:admin_name"`
	Action       string         `json:"action" gorm:"column:action;not null;index" validate:"required,oneof=create update delete approve reject suspend activate process_payout import_winners export_entries login logout change_config manual_adjustment retry_provision"`
	EntityType   string         `json:"entity_type" gorm:"column:entity_type;not null;index" validate:"required,oneof=user draw winner affiliate prize payout system_config admin_user wallet transaction"`
	EntityID     *string        `json:"entity_id" gorm:"column:entity_id;index"`
	Changes      datatypes.JSON `json:"changes" gorm:"column:changes;type:jsonb"`
	IPAddress    *string        `json:"ip_address" gorm:"column:ip_address"`
	UserAgent    *string        `json:"user_agent" gorm:"column:user_agent"`
	Description  *string        `json:"description" gorm:"column:description"`
	Metadata     datatypes.JSON `json:"metadata" gorm:"column:metadata;type:jsonb"`
	Status       string         `json:"status" gorm:"column:status;default:success;not null" validate:"required,oneof=success failed partial"`
	ErrorMessage *string        `json:"error_message" gorm:"column:error_message"`
	CreatedAt    time.Time      `json:"created_at" gorm:"column:created_at;autoCreateTime;index"`
}

// TableName specifies the table name for AuditLog
func (AuditLog) TableName() string {
	return "audit_logs"
}
