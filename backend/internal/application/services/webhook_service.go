package services

import (
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
	log.Printf("[Webhook] ERROR: Invalid webhook signature: %s", signature)
		return errors.Unauthorized("Invalid webhook signature")
	}

	// 2. Parse payload
	var webhookPayload PaystackWebhookPayload
	if err := json.Unmarshal(payload, &webhookPayload); err != nil {
		log.Printf("[Webhook] ERROR: Failed to parse webhook payload: %v", err)
		return errors.BadRequest("Invalid webhook payload")
	}

	log.Printf("[Webhook] Received: Event=%s, Reference=%s, Status=%s, Amount=%d",
		webhookPayload.Event, webhookPayload.Data.Reference, webhookPayload.Data.Status, webhookPayload.Data.Amount)

	// 3. Generate event ID (use Paystack's transaction ID + event type)
	eventID := fmt.Sprintf("%s_%d", webhookPayload.Event, webhookPayload.Data.ID)

	// 4. Check idempotency - has this event been processed before?
	processed, err := s.webhookRepo.IsEventProcessed(ctx, eventID)
	if err != nil {
	log.Printf("[Webhook] ERROR: Failed to check event idempotency for %s: %v", eventID, err)
		return errors.Internal(fmt.Sprintf("Failed to check event status: %v", err))
	}

	if processed {
	log.Printf("[Webhook] Event already processed (idempotent): %s", eventID)
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
	log.Printf("[Webhook] ERROR: Failed to create webhook event %s: %v", eventID, err)
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
	log.Printf("[Webhook] Unhandled event type: %s", webhookPayload.Event)
		// Mark as processed even if we don't handle it
		processErr = nil
	}

	// 7. Update webhook event status
	if processErr != nil {
	log.Printf("[Webhook] ERROR: Failed to process event %s (%s): %v", eventID, webhookPayload.Event, processErr)
		if err := s.webhookRepo.MarkEventFailed(ctx, eventID, processErr.Error()); err != nil {
				log.Printf("[Webhook] ERROR: Failed to mark event as failed: %v", err)
		}
		return processErr
	}

	if err := s.webhookRepo.MarkEventProcessed(ctx, eventID); err != nil {
	log.Printf("[Webhook] ERROR: Failed to mark event %s as processed: %v", eventID, err)
		return errors.Internal(fmt.Sprintf("Failed to update event status: %v", err))
	}

	log.Printf("[Webhook] Processed successfully: EventID=%s, Reference=%s", eventID, webhookPayload.Data.Reference)

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
	log.Printf("[Webhook] WARN: Unknown transaction type - Reference=%s, Prefix=%s", reference, prefix)
		}
	}

	return nil
}

// processChargeFailed handles failed charge events
func (s *WebhookService) processChargeFailed(ctx context.Context, payload *PaystackWebhookPayload) error {
	reference := payload.Data.Reference

	log.Printf("[Webhook] Processing failed charge: Reference=%s, Message=%s", reference, payload.Data.Message)

	// Mark the transaction as FAILED in the database
	if s.rechargeService != nil && reference != "" {
		recharge, err := s.rechargeService.GetRechargeByPaymentRef(ctx, reference)
		if err == nil && recharge != nil && recharge.Status == "PENDING" {
			recharge.Status = "FAILED"
			recharge.FailureReason = fmt.Sprintf("Paystack charge.failed: %s", payload.Data.Message)
			s.rechargeService.UpdateRecharge(ctx, recharge)
			log.Printf("[Webhook] Marked recharge %s as FAILED (ref=%s)", recharge.ID, reference)
		}
	}

	return nil
}
