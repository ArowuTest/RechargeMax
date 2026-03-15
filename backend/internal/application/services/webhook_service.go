package services

import (
	"go.uber.org/zap"
	"rechargemax/internal/logger"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"log"

	"rechargemax/internal/domain/entities"
	"rechargemax/internal/infrastructure/persistence"
	"rechargemax/internal/errors"
)

// PaystackWebhookPayload represents the webhook payload from Paystack
type PaystackWebhookPayload struct {
	Event string `json:"event"`
	Data  struct {
		ID              int64  `json:"id"`
		Domain          string `json:"domain"`
		Status          string `json:"status"`
		Reference       string `json:"reference"`
		Amount          int64  `json:"amount"`
		Message         string `json:"message"`
		GatewayResponse string `json:"gateway_response"`
		PaidAt          string `json:"paid_at"`
		CreatedAt       string `json:"created_at"`
		Channel         string `json:"channel"`
		Currency        string `json:"currency"`
		IPAddress       string `json:"ip_address"`
		Customer        struct {
			ID           int64  `json:"id"`
			Email        string `json:"email"`
			CustomerCode string `json:"customer_code"`
		} `json:"customer"`
	} `json:"data"`
}

// WebhookService handles webhook processing
type WebhookService struct {
	webhookRepo      *persistence.WebhookRepository
	rechargeService  *RechargeService
	subscriptionService *SubscriptionService
	paymentService   *PaymentService
	paystackSecret   string
}

// NewWebhookService creates a new webhook service
func NewWebhookService(
	webhookRepo      *persistence.WebhookRepository,
	rechargeService *RechargeService,
	subscriptionService *SubscriptionService,
	paymentService *PaymentService,
	paystackSecret string,
) *WebhookService {
	return &WebhookService{
		webhookRepo:      webhookRepo,
		rechargeService:  rechargeService,
		subscriptionService: subscriptionService,
		paymentService:   paymentService,
		paystackSecret:   paystackSecret,
	}
}

// VerifyPaystackSignature verifies the webhook signature from Paystack
func (s *WebhookService) VerifyPaystackSignature(payload []byte, signature string) bool {
	mac := hmac.New(sha512.New, []byte(s.paystackSecret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	
	log.Printf("[Webhook] Signature verification - Expected: %s, Received: %s, Match: %v",
		expectedSignature, signature, hmac.Equal([]byte(signature), []byte(expectedSignature)))
	
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// ProcessPaystackWebhook processes a webhook from Paystack
func (s *WebhookService) ProcessPaystackWebhook(ctx context.Context, payload []byte, signature string) error {
	// 1. Verify signature
	if !s.VerifyPaystackSignature(payload, signature) {
	logger.Error("[Webhook] ERROR: Invalid webhook signature", zap.Any("signature", signature))
		return errors.Unauthorized("Invalid webhook signature")
	}

	// 2. Parse payload
	var webhookPayload PaystackWebhookPayload
	if err := json.Unmarshal(payload, &webhookPayload); err != nil {
		logger.Error("[Webhook] ERROR: Failed to parse webhook payload", zap.Error(err))
		return errors.BadRequest("Invalid webhook payload")
	}

	log.Printf("[Webhook] Received: Event=%s, Reference=%s, Status=%s, Amount=%d",
		webhookPayload.Event, webhookPayload.Data.Reference, webhookPayload.Data.Status, webhookPayload.Data.Amount)

	// 3. Generate event ID (use Paystack's transaction ID + event type)
	eventID := fmt.Sprintf("%s_%d", webhookPayload.Event, webhookPayload.Data.ID)

	// 4. Check idempotency - has this event been processed before?
	processed, err := s.webhookRepo.IsEventProcessed(ctx, eventID)
	if err != nil {
	logger.Error("[Webhook] ERROR: Failed to check event idempotency for", zap.Error(err), zap.Any("eventID", eventID))
		return errors.Internal(fmt.Sprintf("Failed to check event status: %v", err))
	}

	if processed {
	logger.Info("[Webhook] Event already processed (idempotent)", zap.Any("eventID", eventID))
		return nil // Not an error, just already processed
	}

	// 5. Create webhook event record
	webhookEvent := &entities.WebhookEvent{
		EventID:    eventID,
		EventType:  webhookPayload.Event,
		Gateway:    "paystack",
		Reference:  webhookPayload.Data.Reference,
		RawPayload: string(payload),
		Signature:  signature,
		Status:     "PENDING",
	}

	if err := s.webhookRepo.CreateEvent(ctx, webhookEvent); err != nil {
	logger.Error("[Webhook] ERROR: Failed to create webhook event", zap.Error(err), zap.Any("eventID", eventID))
		return errors.Internal(fmt.Sprintf("Failed to create webhook event: %v", err))
	}

	// 6. Process based on event type
	var processErr error
	switch webhookPayload.Event {
	case "charge.success":
		processErr = s.processChargeSuccess(ctx, &webhookPayload)
	case "charge.failed":
		processErr = s.processChargeFailed(ctx, &webhookPayload)
	default:
	logger.Info("[Webhook] Unhandled event type", zap.Any("webhookPayload.Event", webhookPayload.Event))
		// Mark as processed even if we don't handle it
		processErr = nil
	}

	// 7. Update webhook event status
	if processErr != nil {
	logger.Error("[Webhook] ERROR: Failed to process event ()", zap.Error(processErr), zap.Any("eventID", eventID), zap.Any("webhookPayload.Event", webhookPayload.Event))
		if err := s.webhookRepo.MarkEventFailed(ctx, eventID, processErr.Error()); err != nil {
				logger.Error("[Webhook] ERROR: Failed to mark event as failed", zap.Error(err))
		}
		return processErr
	}

	if err := s.webhookRepo.MarkEventProcessed(ctx, eventID); err != nil {
	logger.Error("[Webhook] ERROR: Failed to mark event as processed", zap.Error(err), zap.Any("eventID", eventID))
		return errors.Internal(fmt.Sprintf("Failed to update event status: %v", err))
	}

	logger.Info("[Webhook] Processed successfully: EventID=, Reference=", zap.Any("eventID", eventID), zap.Any("webhookPayload.Data.Reference", webhookPayload.Data.Reference))

	return nil
}

// processChargeSuccess handles successful charge events
func (s *WebhookService) processChargeSuccess(ctx context.Context, payload *PaystackWebhookPayload) error {
	reference := payload.Data.Reference

	// Verify payment with Paystack API (double-check)
	success, _, err := s.paymentService.VerifyPayment(ctx, reference, "paystack")
	if err != nil {
		return errors.Internal(fmt.Sprintf("Failed to verify payment: %v", err))
	}

	if !success {
		return errors.BadRequest("Payment verification failed")
	}

	// Determine transaction type from reference prefix
	if len(reference) >= 4 {
		prefix := reference[:4]

		switch prefix {
		case "RCH_":
			// Process recharge
			if err := s.rechargeService.ProcessSuccessfulPayment(ctx, reference); err != nil {
				return errors.Internal(fmt.Sprintf("Failed to process recharge: %v", err))
			}
		case "SUB_":
			// Process subscription
			if err := s.subscriptionService.ProcessSuccessfulPayment(ctx, reference); err != nil {
				return errors.Internal(fmt.Sprintf("Failed to process subscription: %v", err))
			}
		default:
	logger.Warn("[Webhook] WARN: Unknown transaction type - Reference=, Prefix=", zap.Any("reference", reference), zap.Any("prefix", prefix))
		}
	}

	return nil
}

// processChargeFailed handles failed charge events
func (s *WebhookService) processChargeFailed(ctx context.Context, payload *PaystackWebhookPayload) error {
	reference := payload.Data.Reference

	logger.Error("[Webhook] Processing failed charge: Reference=, Message=", zap.Any("reference", reference), zap.Any("payload.Data.Message", payload.Data.Message))

	// Mark the transaction as FAILED in the database
	if s.rechargeService != nil && reference != "" {
		recharge, err := s.rechargeService.GetRechargeByPaymentRef(ctx, reference)
		if err == nil && recharge != nil && recharge.Status == "PENDING" {
			recharge.Status = "FAILED"
			recharge.FailureReason = fmt.Sprintf("Paystack charge.failed: %s", payload.Data.Message)
			s.rechargeService.UpdateRecharge(ctx, recharge)
			logger.Error("[Webhook] Marked recharge as FAILED (ref=)", zap.String("id", recharge.ID.String()), zap.Any("reference", reference))
		}
	}

	return nil
}
