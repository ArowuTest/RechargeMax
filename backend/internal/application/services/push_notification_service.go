package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// PushNotificationService handles FCM push notifications
type PushNotificationService struct {
	fcmServerKey  string
	deviceService *DeviceService
}

// NewPushNotificationService creates a new push notification service
func NewPushNotificationService(fcmServerKey string, deviceService *DeviceService) *PushNotificationService {
	return &PushNotificationService{
		fcmServerKey:  fcmServerKey,
		deviceService: deviceService,
	}
}

// SendToUser sends a push notification to all active devices for the MSISDN.
func (s *PushNotificationService) SendToUser(ctx context.Context, msisdn, title, body string, data map[string]string) error {
	devices, err := s.deviceService.GetDevicesByMSISDN(ctx, msisdn)
	if err != nil {
		return fmt.Errorf("GetDevicesByMSISDN: %w", err)
	}

	var lastErr error
	for _, device := range devices {
		if device.IsActive && device.FCMToken != nil && *device.FCMToken != "" {
			if err := s.SendToToken(ctx, *device.FCMToken, title, body, data); err != nil {
				log.Printf("[push] device %s: %v", device.ID, err)
				lastErr = err
			}
		}
	}
	return lastErr
}

// SendToToken sends a push notification to a specific FCM device token.
// Falls back to stdout logging when no FCM key is configured.
func (s *PushNotificationService) SendToToken(ctx context.Context, fcmToken, title, body string, data map[string]string) error {
	if s.fcmServerKey == "" {
		log.Printf("[PUSH-DEV] Token: %s | %s: %s", fcmToken, title, body)
		return nil
	}

	if data == nil {
		data = map[string]string{}
	}
	payload := map[string]interface{}{
		"to": fcmToken,
		"notification": map[string]string{
			"title": title,
			"body":  body,
		},
		"data":     data,
		"priority": "high",
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal FCM payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://fcm.googleapis.com/fcm/send", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("build FCM request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "key="+s.fcmServerKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("FCM API call: %w", err)
	}
	defer func() {
		io.Copy(io.Discard, resp.Body) //nolint:errcheck
		resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return fmt.Errorf("FCM HTTP %d for token %s", resp.StatusCode, fcmToken)
	}
	return nil
}
