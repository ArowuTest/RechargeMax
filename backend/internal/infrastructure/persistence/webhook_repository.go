package persistence

import (
	"context"
	"time"

	"rechargemax/internal/domain/entities"

	"gorm.io/gorm"
)

// WebhookRepository handles webhook event persistence
type WebhookRepository struct {
	db *gorm.DB
}

// NewWebhookRepository creates a new webhook repository
func NewWebhookRepository(db *gorm.DB) *WebhookRepository {
	return &WebhookRepository{db: db}
}

// IsEventProcessed checks if an event has already been processed (idempotency check)
func (r *WebhookRepository) IsEventProcessed(ctx context.Context, eventID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.WebhookEvent{}).
		Where("event_id = ? AND status = ?", eventID, "PROCESSED").
		Count(&count).Error
	
	return count > 0, err
}

// CreateEvent creates a new webhook event record
func (r *WebhookRepository) CreateEvent(ctx context.Context, event *entities.WebhookEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

// MarkEventProcessed marks an event as successfully processed
func (r *WebhookRepository) MarkEventProcessed(ctx context.Context, eventID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entities.WebhookEvent{}).
		Where("event_id = ?", eventID).
		Updates(map[string]interface{}{
			"status":       "PROCESSED",
			"processed_at": now,
		}).Error
}

// MarkEventFailed marks an event as failed with error message
func (r *WebhookRepository) MarkEventFailed(ctx context.Context, eventID string, errorMsg string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entities.WebhookEvent{}).
		Where("event_id = ?", eventID).
		Updates(map[string]interface{}{
			"status":       "FAILED",
			"processed_at": now,
			"error_msg":    errorMsg,
		}).Error
}

// GetEventByID retrieves a webhook event by its ID
func (r *WebhookRepository) GetEventByID(ctx context.Context, eventID string) (*entities.WebhookEvent, error) {
	var event entities.WebhookEvent
	err := r.db.WithContext(ctx).
		Where("event_id = ?", eventID).
		First(&event).Error
	
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// GetEventsByReference retrieves all webhook events for a payment reference
func (r *WebhookRepository) GetEventsByReference(ctx context.Context, reference string) ([]entities.WebhookEvent, error) {
	var events []entities.WebhookEvent
	err := r.db.WithContext(ctx).
		Where("reference = ?", reference).
		Order("created_at DESC").
		Find(&events).Error
	
	return events, err
}
