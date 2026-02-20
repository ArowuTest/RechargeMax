package persistence

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

// DeviceRepositoryGORM implements the DeviceRepository interface using GORM
type DeviceRepositoryGORM struct {
	db *gorm.DB
}

// NewDeviceRepository creates a new instance of DeviceRepositoryGORM
func NewDeviceRepository(db *gorm.DB) repositories.DeviceRepository {
	return &DeviceRepositoryGORM{db: db}
}

func (r *DeviceRepositoryGORM) Create(ctx context.Context, entity *entities.Device) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

func (r *DeviceRepositoryGORM) FindByID(ctx context.Context, id uuid.UUID) (*entities.Device, error) {
	var entity entities.Device
	err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *DeviceRepositoryGORM) FindAll(ctx context.Context, limit, offset int) ([]*entities.Device, error) {
	var devices []*entities.Device
	err := r.db.WithContext(ctx).
		Order("last_active DESC").
		Limit(limit).
		Offset(offset).
		Find(&devices).Error
	return devices, err
}

func (r *DeviceRepositoryGORM) Update(ctx context.Context, entity *entities.Device) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

func (r *DeviceRepositoryGORM) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.Device{}, "id = ?", id).Error
}

func (r *DeviceRepositoryGORM) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.Device{}).Count(&count).Error
	return count, err
}

// FindByMSISDN retrieves all devices for a specific user
func (r *DeviceRepositoryGORM) FindByMSISDN(ctx context.Context, msisdn string) ([]*entities.Device, error) {
	var devices []*entities.Device
	err := r.db.WithContext(ctx).
		Where("msisdn = ?", msisdn).
		Order("last_active DESC").
		Find(&devices).Error
	return devices, err
}

// FindByDeviceID retrieves a device by its device ID
func (r *DeviceRepositoryGORM) FindByDeviceID(ctx context.Context, deviceID string) (*entities.Device, error) {
	var device entities.Device
	err := r.db.WithContext(ctx).
		Where("device_id = ?", deviceID).
		First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// FindByFCMToken retrieves a device by its FCM token
func (r *DeviceRepositoryGORM) FindByFCMToken(ctx context.Context, fcmToken string) (*entities.Device, error) {
	var device entities.Device
	err := r.db.WithContext(ctx).
		Where("fcm_token = ?", fcmToken).
		First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// FindActiveDevicesByMSISDN retrieves all active devices for a user
func (r *DeviceRepositoryGORM) FindActiveDevicesByMSISDN(ctx context.Context, msisdn string) ([]*entities.Device, error) {
	var devices []*entities.Device
	err := r.db.WithContext(ctx).
		Where("msisdn = ? AND is_active = ?", msisdn, true).
		Order("last_active DESC").
		Find(&devices).Error
	return devices, err
}

// UpdateFCMToken updates the FCM token for a device
func (r *DeviceRepositoryGORM) UpdateFCMToken(ctx context.Context, deviceID string, fcmToken string) error {
	return r.db.WithContext(ctx).
		Model(&entities.Device{}).
		Where("device_id = ?", deviceID).
		Update("fcm_token", fcmToken).
		Error
}

// UpdateLastActive updates the last active timestamp
func (r *DeviceRepositoryGORM) UpdateLastActive(ctx context.Context, deviceID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entities.Device{}).
		Where("device_id = ?", deviceID).
		Update("last_active", now).
		Error
}

// DeactivateDevice marks a device as inactive
func (r *DeviceRepositoryGORM) DeactivateDevice(ctx context.Context, deviceID string) error {
	return r.db.WithContext(ctx).
		Model(&entities.Device{}).
		Where("device_id = ?", deviceID).
		Update("is_active", false).
		Error
}

// IncrementNotificationCount increments the notification count
func (r *DeviceRepositoryGORM) IncrementNotificationCount(ctx context.Context, deviceID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entities.Device{}).
		Where("device_id = ?", deviceID).
		Updates(map[string]interface{}{
			"notification_count":        gorm.Expr("notification_count + 1"),
			"last_notification_sent_at": now,
		}).
		Error
}
