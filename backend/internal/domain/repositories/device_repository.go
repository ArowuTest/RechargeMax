package repositories

import (
	"context"
	"github.com/google/uuid"
	"rechargemax/internal/domain/entities"
)

type DeviceRepository interface {
	// CRUD operations
	Create(ctx context.Context, entity *entities.Device) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Device, error)
	FindAll(ctx context.Context, limit, offset int) ([]*entities.Device, error)
	Update(ctx context.Context, entity *entities.Device) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
	
	// Business operations
	FindByMSISDN(ctx context.Context, msisdn string) ([]*entities.Device, error)
	FindByDeviceID(ctx context.Context, deviceID string) (*entities.Device, error)
	FindByFCMToken(ctx context.Context, fcmToken string) (*entities.Device, error)
	FindActiveDevicesByMSISDN(ctx context.Context, msisdn string) ([]*entities.Device, error)
	UpdateFCMToken(ctx context.Context, deviceID string, fcmToken string) error
	UpdateLastActive(ctx context.Context, deviceID string) error
	DeactivateDevice(ctx context.Context, deviceID string) error
	IncrementNotificationCount(ctx context.Context, deviceID string) error
}
