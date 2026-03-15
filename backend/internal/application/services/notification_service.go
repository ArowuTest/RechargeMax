package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"rechargemax/internal/domain/entities"
	"gorm.io/gorm"
	"rechargemax/internal/domain/repositories"
)

// NotificationService handles multi-channel notifications
type NotificationService struct {
	notificationRepo repositories.NotificationRepository
	deviceRepo       repositories.DeviceRepository
	userRepo         repositories.UserRepository
	smsAPIKey        string
	emailAPIKey      string
	fcmServerKey     string
	db               *gorm.DB
}

// NotificationChannel represents notification delivery channels
type NotificationChannel string

const (
	ChannelSMS   NotificationChannel = "sms"
	ChannelEmail NotificationChannel = "email"
	ChannelPush  NotificationChannel = "push"
	ChannelInApp NotificationChannel = "in_app"
)

// NewNotificationService creates a new notification service
func NewNotificationService(
	notificationRepo repositories.NotificationRepository,
	deviceRepo repositories.DeviceRepository,
	userRepo repositories.UserRepository,
	smsAPIKey, emailAPIKey, fcmServerKey string,
	db *gorm.DB,
) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		deviceRepo:       deviceRepo,
		userRepo:         userRepo,
		smsAPIKey:        smsAPIKey,
		emailAPIKey:      emailAPIKey,
		fcmServerKey:     fcmServerKey,
		db:               db,
	}
}

// logDelivery records a send attempt to notification_delivery_log.
// Called asynchronously — never blocks the calling goroutine.
func (s *NotificationService) logDelivery(channel, status, provider, errorMsg string) {
	if s.db == nil {
		return
	}
	go func() {
		entry := entities.NotificationDeliveryLog{
			Channel:        channel,
			DeliveryStatus: status,
			ProviderName:   provider,
			ErrorMessage:   errorMsg,
		}
		s.db.Create(&entry) // best-effort; ignore errors
	}()
}

// SendSMS sends SMS notification via Termii or similar service
func (s *NotificationService) SendSMS(ctx context.Context, msisdn, message string) error {
	// Integrate with Termii SMS API
	// In production, this would:
	// 1. Make HTTP POST request to Termii API
	// 2. Include API key in headers
	// 3. Send SMS to Nigerian number
	// 4. Handle response and errors
	//
	// Example implementation:
	// import "bytes"
	// import "net/http"
	// 
	// payload := map[string]interface{}{
	//     "to":      msisdn,
	//     "from":    "RechargeMax",
	//     "sms":     message,
	//     "type":    "plain",
	//     "channel": "generic",
	//     "api_key": s.smsAPIKey,
	// }
	// 
	// jsonData, _ := json.Marshal(payload)
	// req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.ng.termii.com/api/sms/send", bytes.NewBuffer(jsonData))
	// req.Header.Set("Content-Type", "application/json")
	// 
	// client := &http.Client{Timeout: 10 * time.Second}
	// resp, err := client.Do(req)
	// if err != nil {
	//     return fmt.Errorf("failed to send SMS: %w", err)
	// }
	// defer resp.Body.Close()
	// 
	// if resp.StatusCode != 200 {
	//     return fmt.Errorf("SMS API returned status %d", resp.StatusCode)
	// }
	
	// For now, log the SMS (when Termii API key is configured, uncomment above)
	if s.smsAPIKey != "" {
		fmt.Printf("[SMS] To: %s, Message: %s\n", msisdn, message)
		// Actual API call would go here
		s.logDelivery("sms", "sent", "termii", "")
	} else {
		fmt.Printf("[SMS-MOCK] To: %s, Message: %s\n", msisdn, message)
		s.logDelivery("sms", "sent", "mock", "")
	}
	return nil
}

// SendEmail sends email notification via SendGrid or similar service
func (s *NotificationService) SendEmail(ctx context.Context, email, subject, body string) error {
	// Integrate with SendGrid Email API
	// In production, this would:
	// 1. Use SendGrid Go SDK or HTTP API
	// 2. Create email with from/to/subject/body
	// 3. Send via SendGrid API
	// 4. Handle response and errors
	//
	// Example implementation using SendGrid SDK:
	// import "github.com/sendgrid/sendgrid-go"
	// import "github.com/sendgrid/sendgrid-go/helpers/mail"
	// 
	// from := mail.NewEmail("RechargeMax", "noreply@rechargemax.ng")
	// to := mail.NewEmail("", email)
	// message := mail.NewSingleEmail(from, subject, to, body, body)
	// 
	// client := sendgrid.NewSendClient(s.emailAPIKey)
	// response, err := client.Send(message)
	// if err != nil {
	//     return fmt.Errorf("failed to send email: %w", err)
	// }
	// 
	// if response.StatusCode >= 400 {
	//     return fmt.Errorf("email API returned status %d", response.StatusCode)
	// }
	
	// For now, log the email (when SendGrid API key is configured, uncomment above)
	if s.emailAPIKey != "" {
		fmt.Printf("[EMAIL] To: %s, Subject: %s, Body: %s\n", email, subject, body)
		// Actual API call would go here
	} else {
		fmt.Printf("[EMAIL-MOCK] To: %s, Subject: %s, Body: %s\n", email, subject, body)
	}
	return nil
}

// SendPushNotification sends push notification via FCM
func (s *NotificationService) SendPushNotification(ctx context.Context, msisdn, title, body string) error {
	// Get user's devices
	devices, err := s.deviceRepo.FindByMSISDN(ctx, msisdn)
	if err != nil || len(devices) == 0 {
		return nil // No devices registered, skip push notification
	}

	// Integrate with FCM (Firebase Cloud Messaging)
	// In production, this would:
	// 1. Use Firebase Admin SDK or HTTP API
	// 2. Send push notification to each device token
	// 3. Handle invalid tokens (remove from database)
	// 4. Track delivery status
	//
	// Example implementation using FCM HTTP API:
	// import "bytes"
	// import "net/http"
	// 
	// for _, device := range devices {
	//     if device.FCMToken == nil || *device.FCMToken == "" {
	//         continue
	//     }
	//     
	//     payload := map[string]interface{}{
	//         "to": *device.FCMToken,
	//         "notification": map[string]string{
	//             "title": title,
	//             "body":  body,
	//         },
	//         "data": map[string]string{
	//             "msisdn": msisdn,
	//         },
	//     }
	//     
	//     jsonData, _ := json.Marshal(payload)
	//     req, _ := http.NewRequestWithContext(ctx, "POST", "https://fcm.googleapis.com/fcm/send", bytes.NewBuffer(jsonData))
	//     req.Header.Set("Content-Type", "application/json")
	//     req.Header.Set("Authorization", "key="+s.fcmServerKey)
	//     
	//     client := &http.Client{Timeout: 10 * time.Second}
	//     resp, err := client.Do(req)
	//     if err != nil {
	//         fmt.Printf("Failed to send push to %s: %v\n", *device.FCMToken, err)
	//         continue
	//     }
	//     defer resp.Body.Close()
	//     
	//     if resp.StatusCode != 200 {
	//         fmt.Printf("FCM API returned status %d for token %s\n", resp.StatusCode, *device.FCMToken)
	//     }
	// }
	
	// For now, log the push notifications (when FCM server key is configured, uncomment above)
	for _, device := range devices {
		token := "(no token)"
		if device.FCMToken != nil {
			token = *device.FCMToken
		}
		if s.fcmServerKey != "" {
			fmt.Printf("[PUSH] To: %s (Device: %s), Title: %s, Body: %s\n", msisdn, token, title, body)
			// Actual API call would go here
		} else {
			fmt.Printf("[PUSH-MOCK] To: %s (Device: %s), Title: %s, Body: %s\n", msisdn, token, title, body)
		}
	}

	return nil
}

// CreateNotification creates an in-platform notification
func (s *NotificationService) CreateNotification(ctx context.Context, msisdn, notificationType, title, message string, metadata map[string]interface{}) error {
	metadataJSON, _ := json.Marshal(metadata)
	notification := &entities.Notification{
		ID:       uuid.New(),
		MSISDN:   msisdn,
		Type:     notificationType,
		Title:    title,
		Message:  message,
		Priority: "normal",
		Metadata: datatypes.JSON(metadataJSON),
		IsRead:   false,
	}

	return s.notificationRepo.Create(ctx, notification)
}

// SendMultiChannel sends notification via all available channels
func (s *NotificationService) SendMultiChannel(ctx context.Context, msisdn, title, message string, notificationType string, metadata map[string]interface{}) error {
	// Get user details
	user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 1. Send SMS
	s.SendSMS(ctx, msisdn, message)

	// 2. Send Email (if available)
	if user.Email != "" {
		s.SendEmail(ctx, user.Email, title, message)
	}

	// 3. Send Push Notification
	s.SendPushNotification(ctx, msisdn, title, message)

	// 4. Create In-Platform Notification
	s.CreateNotification(ctx, msisdn, notificationType, title, message, metadata)

	return nil
}

// GetNotifications retrieves user notifications with pagination
func (s *NotificationService) GetNotifications(ctx context.Context, msisdn string, page, limit int) ([]*entities.Notification, int64, error) {
	notifications, err := s.notificationRepo.FindByMSISDN(ctx, msisdn, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get notifications: %w", err)
	}

	total, err := s.notificationRepo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count notifications: %w", err)
	}

	return notifications, total, nil
}

// GetUnreadNotifications retrieves unread notifications
func (s *NotificationService) GetUnreadNotifications(ctx context.Context, msisdn string) ([]*entities.Notification, error) {
	return s.notificationRepo.FindUnreadByMSISDN(ctx, msisdn, 100, 0)
}

// GetUnreadCount gets count of unread notifications
func (s *NotificationService) GetUnreadCount(ctx context.Context, msisdn string) (int64, error) {
	return s.notificationRepo.CountUnreadByMSISDN(ctx, msisdn)
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(ctx context.Context, notificationID uuid.UUID) error {
	notification, err := s.notificationRepo.FindByID(ctx, notificationID)
	if err != nil {
		return fmt.Errorf("notification not found: %w", err)
	}

	notification.IsRead = true
	now := time.Now()
	notification.ReadAt = &now

	return s.notificationRepo.Update(ctx, notification)
}

// MarkAllAsRead marks all notifications as read for a user
func (s *NotificationService) MarkAllAsRead(ctx context.Context, msisdn string) error {
	notifications, err := s.GetUnreadNotifications(ctx, msisdn)
	if err != nil {
		return err
	}

	for _, notif := range notifications {
		notif.IsRead = true
		now := time.Now()
		notif.ReadAt = &now
		s.notificationRepo.Update(ctx, notif)
	}

	return nil
}

// DeleteNotification deletes a notification
func (s *NotificationService) DeleteNotification(ctx context.Context, notificationID uuid.UUID) error {
	return s.notificationRepo.Delete(ctx, notificationID)
}

// SendWelcomeNotification sends welcome notification to new users
func (s *NotificationService) SendWelcomeNotification(ctx context.Context, msisdn string) error {
	title := "Welcome to RechargeMax! 🎉"
	message := "Thank you for joining RechargeMax! Recharge your phone and earn rewards. Spin the wheel, win prizes, and refer friends to earn commissions!"

	return s.SendMultiChannel(ctx, msisdn, title, message, "system", nil)
}

// SendRechargeSuccessNotification sends notification after successful recharge
func (s *NotificationService) SendRechargeSuccessNotification(ctx context.Context, msisdn string, amount int64, network string, pointsEarned int64) error {
	title := "Recharge Successful! ✅"
	message := fmt.Sprintf("Your ₦%d %s recharge was successful! You earned %d points.", amount/100, network, pointsEarned)

	return s.SendMultiChannel(ctx, msisdn, title, message, "system", map[string]interface{}{
		"amount":        amount,
		"network":       network,
		"points_earned": pointsEarned,
	})
}

// SendSpinEligibilityNotification sends notification when user becomes eligible to spin
func (s *NotificationService) SendSpinEligibilityNotification(ctx context.Context, msisdn string) error {
	title := "Spin the Wheel! 🎰"
	message := "You're eligible to spin the wheel! Login now to win data, airtime, or bonus points!"

	return s.SendMultiChannel(ctx, msisdn, title, message, "spin_win", nil)
}

// SendSpinWinNotification sends notification after winning a spin
func (s *NotificationService) SendSpinWinNotification(ctx context.Context, msisdn, prizeType string, prizeValue int64) error {
	title := "You Won! 🎉"
	var message string

	if prizeType == "data" {
		message = fmt.Sprintf("Congratulations! You won %dGB data! It has been credited to your number.", prizeValue)
	} else if prizeType == "airtime" {
		message = fmt.Sprintf("Congratulations! You won ₦%d airtime! It has been credited to your number.", prizeValue/100)
	} else if prizeType == "points" {
		message = fmt.Sprintf("Congratulations! You won %d bonus points!", prizeValue)
	}

	return s.SendMultiChannel(ctx, msisdn, title, message, "spin_win", map[string]interface{}{
		"prize_type":  prizeType,
		"prize_value": prizeValue,
	})
}

// SendReferralSignupNotification sends notification when a referral signs up
func (s *NotificationService) SendReferralSignupNotification(ctx context.Context, affiliateMSISDN, referredMSISDN string) error {
	title := "New Referral! 👥"
	message := fmt.Sprintf("Great news! %s signed up using your referral link!", referredMSISDN)

	return s.SendMultiChannel(ctx, affiliateMSISDN, title, message, "system", map[string]interface{}{
		"referred_msisdn": referredMSISDN,
	})
}

// SendCommissionEarnedNotification sends notification when commission is earned
func (s *NotificationService) SendCommissionEarnedNotification(ctx context.Context, affiliateMSISDN string, commissionAmount int64, referredMSISDN string) error {
	title := "Commission Earned! 💰"
	message := fmt.Sprintf("You earned ₦%.2f commission from %s's recharge! It will be available in 7 days.", float64(commissionAmount)/100, referredMSISDN)

	return s.SendMultiChannel(ctx, affiliateMSISDN, title, message, "commission_earned", map[string]interface{}{
		"commission_amount": commissionAmount,
		"referred_msisdn":   referredMSISDN,
	})
}

// SendPayoutSuccessNotification sends notification after successful payout
func (s *NotificationService) SendPayoutSuccessNotification(ctx context.Context, msisdn string, amount int64, accountNumber string) error {
	title := "Payout Successful! ✅"
	message := fmt.Sprintf("Your payout of ₦%.2f has been sent to account %s!", float64(amount)/100, accountNumber)

	return s.SendMultiChannel(ctx, msisdn, title, message, "system", map[string]interface{}{
		"amount":         amount,
		"account_number": accountNumber,
	})
}

// SendSubscriptionRenewalNotification sends notification for subscription renewal
func (s *NotificationService) SendSubscriptionRenewalNotification(ctx context.Context, msisdn string, amount int64) error {
	title := "Subscription Renewed! 🔄"
	message := fmt.Sprintf("Your ₦%d daily subscription has been renewed. You earned 1 draw entry!", amount/100)

	return s.SendMultiChannel(ctx, msisdn, title, message, "system", map[string]interface{}{
		"amount": amount,
	})
}

// SendSubscriptionCancelledNotification sends notification when subscription is cancelled
func (s *NotificationService) SendSubscriptionCancelledNotification(ctx context.Context, msisdn string) error {
	title := "Subscription Cancelled"
	message := "Your daily subscription has been cancelled. You can resubscribe anytime to continue earning rewards!"

	return s.SendMultiChannel(ctx, msisdn, title, message, "system", nil)
}
