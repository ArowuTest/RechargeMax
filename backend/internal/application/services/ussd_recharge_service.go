package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

// USSDRechargeService handles USSD recharge tracking and points allocation
type USSDRechargeService struct {
	ussdRepo repositories.USSDRechargeRepository
	userRepo repositories.UserRepository
	notificationService *NotificationService
}

// NewUSSDRechargeService creates a new USSD recharge service
func NewUSSDRechargeService(
	ussdRepo repositories.USSDRechargeRepository,
	userRepo repositories.UserRepository,
	notificationService *NotificationService,
) *USSDRechargeService {
	return &USSDRechargeService{
		ussdRepo: ussdRepo,
		userRepo: userRepo,
		notificationService: notificationService,
	}
}

// USSDWebhookPayload represents the webhook payload from telecom providers
type USSDWebhookPayload struct {
	TransactionRef string    `json:"transaction_ref"`
	ProviderRef    string    `json:"provider_ref"`
	MSISDN         string    `json:"msisdn"`
	Network        string    `json:"network"`
	Amount         int64     `json:"amount"` // Amount in kobo
	RechargeType   string    `json:"recharge_type"` // airtime, data
	ProductCode    string    `json:"product_code"`
	Status         string    `json:"status"`
	RechargeDate   time.Time `json:"recharge_date"`
}

// ProcessWebhook processes incoming USSD recharge webhook
func (s *USSDRechargeService) ProcessWebhook(ctx context.Context, provider, endpoint, method string, headers, body, ipAddress string) error {
	// Create webhook log
	webhookLog := &entities.USSDWebhookLog{
		ID:         uuid.New(),
		Provider:   provider,
		Endpoint:   endpoint,
		Method:     method,
		Headers:    headers,
		Body:       body,
		IPAddress:  ipAddress,
		Status:     "received",
		ReceivedAt: time.Now(),
	}

	if err := s.ussdRepo.CreateWebhookLog(ctx, webhookLog); err != nil {
		return fmt.Errorf("failed to create webhook log: %w", err)
	}

	// Parse payload
	var payload USSDWebhookPayload
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		webhookLog.Status = "failed"
		webhookLog.ProcessingError = fmt.Sprintf("failed to parse payload: %v", err)
		s.ussdRepo.UpdateWebhookLog(ctx, webhookLog)
		return fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	// Check for duplicate transaction
	existing, err := s.ussdRepo.FindByTransactionRef(ctx, payload.TransactionRef)
	if err == nil && existing != nil {
		webhookLog.Status = "processed"
		webhookLog.USSDRechargeID = &existing.ID
		s.ussdRepo.UpdateWebhookLog(ctx, webhookLog)
		return nil // Already processed
	}

	// Create USSD recharge record
	ussdRecharge, err := s.createUSSDRecharge(ctx, payload, body)
	if err != nil {
		webhookLog.Status = "failed"
		webhookLog.ProcessingError = fmt.Sprintf("failed to create recharge: %v", err)
		s.ussdRepo.UpdateWebhookLog(ctx, webhookLog)
		return err
	}

	// Update webhook log
	now := time.Now()
	webhookLog.Status = "processed"
	webhookLog.USSDRechargeID = &ussdRecharge.ID
	webhookLog.ProcessedAt = &now
	s.ussdRepo.UpdateWebhookLog(ctx, webhookLog)

	return nil
}

func (s *USSDRechargeService) createUSSDRecharge(ctx context.Context, payload USSDWebhookPayload, rawPayload string) (*entities.USSDRecharge, error) {
	// Convert amount to kobo
	amountInKobo := int64(payload.Amount * 100)

	// Calculate points (₦200 = 1 point)
	pointsEarned := int(amountInKobo / 20000)

	// Get or create user
	user, err := s.userRepo.FindByMSISDN(ctx, payload.MSISDN)
	var userID *uuid.UUID
	if err == nil && user != nil {
		userID = &user.ID
	}

	// Create USSD recharge record
	ussdRecharge := &entities.USSDRecharge{
		ID:             uuid.New(),
		UserID:         userID,
		MSISDN:         payload.MSISDN,
		Network:        payload.Network,
		Amount:         amountInKobo,
		RechargeType:   payload.RechargeType,
		ProductCode:    payload.ProductCode,
		TransactionRef: payload.TransactionRef,
		ProviderRef:    payload.ProviderRef,
		PointsEarned:   pointsEarned,
		Status:         payload.Status,
		RechargeDate:   payload.RechargeDate,
		ReceivedAt:     time.Now(),
		WebhookPayload: rawPayload,
	}

	if err := s.ussdRepo.Create(ctx, ussdRecharge); err != nil {
		return nil, fmt.Errorf("failed to create USSD recharge: %w", err)
	}

	// Allocate points to user
	if err := s.allocatePoints(ctx, ussdRecharge); err != nil {
		return nil, fmt.Errorf("failed to allocate points: %w", err)
	}

	// TODO: Implement SendUSSDRechargeNotification
	// s.notificationService.SendUSSDRechargeNotification(ctx, payload.MSISDN, amountInKobo, pointsEarned)

	return ussdRecharge, nil
}

func (s *USSDRechargeService) allocatePoints(ctx context.Context, ussdRecharge *entities.USSDRecharge) error {
	// Get user
	user, err := s.userRepo.FindByMSISDN(ctx, ussdRecharge.MSISDN)
	if err != nil {
		// User doesn't exist yet, create basic user record
		user = &entities.Users{
			ID:          uuid.New(),
			MSISDN:      ussdRecharge.MSISDN,
			TotalPoints: ussdRecharge.PointsEarned,
			CreatedAt:   time.Now(),
		}
		if err := s.userRepo.Create(ctx, user); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
	} else {
		// Update existing user points
		user.TotalPoints += ussdRecharge.PointsEarned
		if err := s.userRepo.Update(ctx, user); err != nil {
			return fmt.Errorf("failed to update user points: %w", err)
		}
	}

	// Update USSD recharge as processed
	now := time.Now()
	ussdRecharge.ProcessedAt = &now
	ussdRecharge.UserID = &user.ID

	if err := s.ussdRepo.Update(ctx, ussdRecharge); err != nil {
		return fmt.Errorf("failed to update USSD recharge: %w", err)
	}

	return nil
}

// GetUSSDRechargesByMSISDN returns USSD recharges for a specific MSISDN
func (s *USSDRechargeService) GetUSSDRechargesByMSISDN(ctx context.Context, msisdn string, startDate, endDate time.Time) ([]*entities.USSDRecharge, error) {
	return s.ussdRepo.FindByMSISDN(ctx, msisdn, startDate, endDate)
}

// ProcessUnprocessedRecharges processes any USSD recharges that failed to allocate points
func (s *USSDRechargeService) ProcessUnprocessedRecharges(ctx context.Context) error {
	recharges, err := s.ussdRepo.FindUnprocessed(ctx)
	if err != nil {
		return fmt.Errorf("failed to find unprocessed recharges: %w", err)
	}

	for _, recharge := range recharges {
		if err := s.allocatePoints(ctx, recharge); err != nil {
			fmt.Printf("Failed to allocate points for USSD recharge %s: %v\n", recharge.ID, err)
		}
	}

	return nil
}

// GetWebhookLogs returns webhook logs for debugging
func (s *USSDRechargeService) GetWebhookLogs(ctx context.Context, provider string, startDate, endDate time.Time) ([]*entities.USSDWebhookLog, error) {
	return s.ussdRepo.FindWebhookLogs(ctx, provider, startDate, endDate)
}

// RetryFailedWebhooks retries processing of failed webhooks
func (s *USSDRechargeService) RetryFailedWebhooks(ctx context.Context) error {
	logs, err := s.ussdRepo.FindFailedWebhookLogs(ctx)
	if err != nil {
		return fmt.Errorf("failed to find failed webhooks: %w", err)
	}

	for _, log := range logs {
		// Re-process webhook
		if err := s.ProcessWebhook(ctx, log.Provider, log.Endpoint, log.Method, log.Headers, log.Body, log.IPAddress); err != nil {
			fmt.Printf("Failed to retry webhook %s: %v\n", log.ID, err)
		}
	}

	return nil
}
