package entities

import (
	"time"

	"github.com/google/uuid"
)

// WebhookEvent represents a webhook event received from payment gateway
type WebhookEvent struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EventID     string    `gorm:"uniqueIndex;not null" json:"event_id"`           // Unique event ID from payment gateway
	EventType   string    `gorm:"not null" json:"event_type"`                     // e.g., "charge.success"
	Gateway     string    `gorm:"not null;default:'paystack'" json:"gateway"`     // Payment gateway name
	Reference   string    `gorm:"index;not null" json:"reference"`                // Payment reference
	RawPayload  string    `gorm:"type:text" json:"raw_payload"`                   // Full webhook payload (encrypted)
	Signature   string    `gorm:"type:text" json:"signature"`                     // Webhook signature for verification
	Status      string    `gorm:"not null;default:'PENDING'" json:"status"`       // PENDING, PROCESSED, FAILED
	ProcessedAt *time.Time `json:"processed_at"`                                  // When the event was processed
	ErrorMsg    string    `gorm:"type:text" json:"error_msg,omitempty"`           // Error message if processing failed
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for WebhookEvent
func (WebhookEvent) TableName() string {
	return "webhook_events"
}
