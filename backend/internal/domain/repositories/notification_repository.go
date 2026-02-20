package repositories

import (
	"context"
	"github.com/google/uuid"
	"rechargemax/internal/domain/entities"
)

type NotificationRepository interface {
	// CRUD operations
	Create(ctx context.Context, entity *entities.Notification) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Notification, error)
	FindAll(ctx context.Context, limit, offset int) ([]*entities.Notification, error)
	Update(ctx context.Context, entity *entities.Notification) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
	
	// Business operations
	FindByMSISDN(ctx context.Context, msisdn string, limit, offset int) ([]*entities.Notification, error)
	FindUnreadByMSISDN(ctx context.Context, msisdn string, limit, offset int) ([]*entities.Notification, error)
	CountUnreadByMSISDN(ctx context.Context, msisdn string) (int64, error)
	MarkAsRead(ctx context.Context, notificationID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, msisdn string) error
	FindByType(ctx context.Context, notifType string, limit, offset int) ([]*entities.Notification, error)
	FindByReference(ctx context.Context, referenceType, referenceID string) ([]*entities.Notification, error)
	DeleteOldNotifications(ctx context.Context, daysOld int) (int64, error)
}
