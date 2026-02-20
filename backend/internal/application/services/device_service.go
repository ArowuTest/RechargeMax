package services

import (
	"context"

	"github.com/google/uuid"
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

// DeviceService handles device management
type DeviceService struct {
	deviceRepo repositories.DeviceRepository
}

// NewDeviceService creates a new device service
func NewDeviceService(deviceRepo repositories.DeviceRepository) *DeviceService {
	return &DeviceService{
		deviceRepo: deviceRepo,
	}
}

// RegisterDevice registers a device for push notifications
func (s *DeviceService) RegisterDevice(ctx context.Context, msisdn, deviceID, fcmToken, platform string) error {
	fcmTokenPtr := &fcmToken
	device := &entities.Device{
		ID:       uuid.New(),
		MSISDN:   msisdn,
		DeviceID: deviceID,
		FCMToken: fcmTokenPtr,
		Platform: platform,
		IsActive: true,
	}
	return s.deviceRepo.Create(ctx, device)
}

// UnregisterDevice unregisters a device by device ID
func (s *DeviceService) UnregisterDevice(ctx context.Context, deviceID string) error {
	// Find device by device_id
	device, err := s.deviceRepo.FindByDeviceID(ctx, deviceID)
	if err != nil {
		return err
	}
	return s.deviceRepo.Delete(ctx, device.ID)
}

// GetDevicesByMSISDN gets all devices for a user
func (s *DeviceService) GetDevicesByMSISDN(ctx context.Context, msisdn string) ([]*entities.Device, error) {
	return s.deviceRepo.FindByMSISDN(ctx, msisdn)
}
