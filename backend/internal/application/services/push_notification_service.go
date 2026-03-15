package services

import (
	"context"
	"fmt"
	"log"
)

// PushNotificationService handles FCM push notifications
type PushNotificationService struct {
	fcmServerKey string
	deviceService *DeviceService
}

// NewPushNotificationService creates a new push notification service
func NewPushNotificationService(fcmServerKey string, deviceService *DeviceService) *PushNotificationService {
	return &PushNotificationService{
		fcmServerKey:  fcmServerKey,
		deviceService: deviceService,
	}
}

// SendToUser sends a push notification to all user's devices
func (s *PushNotificationService) SendToUser(ctx context.Context, msisdn, title, body string, data map[string]string) error {
	devices, err := s.deviceService.GetDevicesByMSISDN(ctx, msisdn)
	if err != nil {
		return fmt.Errorf("failed to get devices: %w", err)
	}

		for _, device := range devices {
			if device.IsActive && device.FCMToken != nil {
				// Implement actual FCM API call
				// In production, this would:
				// 1. Make HTTP POST to FCM API
				// 2. Include FCM server key in Authorization header
				// 3. Send notification payload
				// 4. Handle errors (invalid token, etc.)
				//
				// Example:
				// err := s.SendToToken(ctx, *device.FCMToken, title, body, data)
				// if err != nil {
				//     log.Printf("Failed to send push to device %s: %v\n", device.ID, err)
				// }
				
				// For now, log the push notification
				log.Printf("[PUSH] To device %s (Token: %s), Title: %s, Body: %s\n", device.ID, *device.FCMToken, title, body)
			}
		}

	return nil
}

// SendToToken sends a push notification to a specific device token
func (s *PushNotificationService) SendToToken(ctx context.Context, fcmToken, title, body string, data map[string]string) error {
	// Implement actual FCM API call
	// In production, this would:
	// 1. Create FCM payload with notification and data
	// 2. Make HTTP POST to https://fcm.googleapis.com/fcm/send
	// 3. Include Authorization: key=<fcmServerKey> header
	// 4. Handle response and errors
	//
	// Example implementation:
	// import "bytes"
	// import "encoding/json"
	// import "net/http"
	// 
	// payload := map[string]interface{}{
	//     "to": fcmToken,
	//     "notification": map[string]string{
	//         "title": title,
	//         "body":  body,
	//     },
	//     "data": data,
	// }
	// 
	// jsonData, _ := json.Marshal(payload)
	// req, _ := http.NewRequestWithContext(ctx, "POST", "https://fcm.googleapis.com/fcm/send", bytes.NewBuffer(jsonData))
	// req.Header.Set("Content-Type", "application/json")
	// req.Header.Set("Authorization", "key="+s.fcmServerKey)
	// 
	// client := &http.Client{Timeout: 10 * time.Second}
	// resp, err := client.Do(req)
	// if err != nil {
	//     return fmt.Errorf("failed to send FCM notification: %w", err)
	// }
	// defer resp.Body.Close()
	// 
	// if resp.StatusCode != 200 {
	//     return fmt.Errorf("FCM API returned status %d", resp.StatusCode)
	// }
	
	// For now, log the push notification (when FCM server key is configured, uncomment above)
	log.Printf("[PUSH-TOKEN] To: %s, Title: %s, Body: %s\n", fcmToken, title, body)
	return nil
}
