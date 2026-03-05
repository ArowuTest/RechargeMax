package entities

import (
	"time"

	"github.com/google/uuid"
)

// DrawEntries represents the draw_entries table
type DrawEntries struct {
	ID           uuid.UUID  `json:"id" gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	DrawID       uuid.UUID  `json:"draw_id" gorm:"column:draw_id;not null;index"`
	UserID       *uuid.UUID `json:"user_id" gorm:"column:user_id;index"`
	Msisdn       string     `json:"msisdn" gorm:"column:msisdn;not null" validate:"required"`
	EntriesCount         *int       `json:"entries_count" gorm:"column:entries_count"`
	SourceType           string     `json:"source_type" gorm:"column:source_type;default:TRANSACTION"` // TRANSACTION, SUBSCRIPTION, BONUS, MANUAL
	SourceTransactionID  *uuid.UUID `json:"source_transaction_id" gorm:"column:source_transaction_id"`
	SourceSubscriptionID *uuid.UUID `json:"source_subscription_id" gorm:"column:source_subscription_id"`
	CreatedAt            *time.Time `json:"created_at" gorm:"column:created_at"`
}

// TableName specifies the table name for DrawEntries
func (DrawEntries) TableName() string {
	return "draw_entries"
}
