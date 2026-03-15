package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

// SubscriptionTierService handles subscription tier and pricing management
type SubscriptionTierService struct {
	tierRepo    repositories.SubscriptionTierRepository
	userRepo    repositories.UserRepository
	paymentService *PaymentService
	notificationService *NotificationService
}

// NewSubscriptionTierService creates a new subscription tier service
func NewSubscriptionTierService(
	tierRepo repositories.SubscriptionTierRepository,
	userRepo repositories.UserRepository,
	paymentService *PaymentService,
	notificationService *NotificationService,
) *SubscriptionTierService {
	return &SubscriptionTierService{
		tierRepo:    tierRepo,
		userRepo:    userRepo,
		paymentService: paymentService,
		notificationService: notificationService,
	}
}

// CreateTier creates a new subscription tier
func (s *SubscriptionTierService) CreateTier(ctx context.Context, name, description string, entries, sortOrder int) (*entities.SubscriptionTier, error) {
	tier := &entities.SubscriptionTier{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Entries:     entries,
		IsActive:    true,
		SortOrder:   sortOrder,
	}

	if err := s.tierRepo.Create(ctx, tier); err != nil {
		return nil, fmt.Errorf("failed to create tier: %w", err)
	}

	return tier, nil
}

// UpdateTier updates an existing subscription tier
func (s *SubscriptionTierService) UpdateTier(ctx context.Context, id uuid.UUID, name, description string, entries, sortOrder int, isActive bool) (*entities.SubscriptionTier, error) {
	tier, err := s.tierRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("tier not found: %w", err)
	}

	tier.Name = name
	tier.Description = description
	tier.Entries = entries
	tier.SortOrder = sortOrder
	tier.IsActive = isActive

	if err := s.tierRepo.Update(ctx, tier); err != nil {
		return nil, fmt.Errorf("failed to update tier: %w", err)
	}

	return tier, nil
}

// DeleteTier soft deletes a subscription tier
func (s *SubscriptionTierService) DeleteTier(ctx context.Context, id uuid.UUID) error {
	return s.tierRepo.Delete(ctx, id)
}

// GetAllTiers returns all subscription tiers
func (s *SubscriptionTierService) GetAllTiers(ctx context.Context) ([]*entities.SubscriptionTier, error) {
	return s.tierRepo.FindAll(ctx)
}

// GetActiveTiers returns only active subscription tiers
func (s *SubscriptionTierService) GetActiveTiers(ctx context.Context) ([]*entities.SubscriptionTier, error) {
	return s.tierRepo.FindActive(ctx)
}

// SetPricePerEntry sets the global price per entry
func (s *SubscriptionTierService) SetPricePerEntry(ctx context.Context, priceInKobo int64) (*entities.SubscriptionPricing, error) {
	// Deactivate current pricing
	currentPricing, err := s.tierRepo.GetCurrentPricing(ctx)
	if err == nil && currentPricing != nil {
		now := time.Now()
		currentPricing.EffectiveTo = &now
		currentPricing.IsActive = false
		if err := s.tierRepo.UpdatePricing(ctx, currentPricing); err != nil {
			return nil, fmt.Errorf("failed to deactivate current pricing: %w", err)
		}
	}

	// Create new pricing
	newPricing := &entities.SubscriptionPricing{
		ID:            uuid.New(),
		PricePerEntry: priceInKobo,
		Currency:      "NGN",
		IsActive:      true,
		EffectiveFrom: time.Now(),
	}

	if err := s.tierRepo.CreatePricing(ctx, newPricing); err != nil {
		return nil, fmt.Errorf("failed to create new pricing: %w", err)
	}

	return newPricing, nil
}

// GetCurrentPricing returns the current active pricing
func (s *SubscriptionTierService) GetCurrentPricing(ctx context.Context) (*entities.SubscriptionPricing, error) {
	return s.tierRepo.GetCurrentPricing(ctx)
}

// CalculateSubscriptionCost calculates the total cost for a subscription
func (s *SubscriptionTierService) CalculateSubscriptionCost(ctx context.Context, tierID uuid.UUID, bundleQuantity int) (int64, int, error) {
	// Get tier
	tier, err := s.tierRepo.FindByID(ctx, tierID)
	if err != nil {
		return 0, 0, fmt.Errorf("tier not found: %w", err)
	}

	if !tier.IsActive {
		return 0, 0, fmt.Errorf("tier is not active")
	}

	// Get current pricing
	pricing, err := s.tierRepo.GetCurrentPricing(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("pricing not found: %w", err)
	}

	// Calculate cost and total entries
	totalEntries := tier.Entries * bundleQuantity
	totalCost := pricing.PricePerEntry * int64(totalEntries)

	return totalCost, totalEntries, nil
}

// CreateDailySubscription creates a new daily subscription with bundle quantity
func (s *SubscriptionTierService) CreateDailySubscription(ctx context.Context, msisdn string, tierID uuid.UUID, bundleQuantity int, paymentMethod string) (*entities.DailySubscription, string, error) {
	// Calculate cost
	totalCost, totalEntries, err := s.CalculateSubscriptionCost(ctx, tierID, bundleQuantity)
	if err != nil {
		return nil, "", err
	}

	// Check for existing active subscription
	existingSubs, err := s.tierRepo.FindDailySubscriptionsByMSISDN(ctx, msisdn)
	if err == nil {
		for _, sub := range existingSubs {
			if sub.Status == "active" {
				return nil, "", fmt.Errorf("user already has an active subscription")
			}
		}
	}

	// Get or create user
	user, err := s.userRepo.FindByMSISDN(ctx, msisdn)
	var userID *uuid.UUID
	if err == nil && user != nil {
		userID = &user.ID
	}

	// Create subscription
	subscription := &entities.DailySubscription{
		ID:              uuid.New(),
		UserID:          userID,
		MSISDN:          msisdn,
		TierID:          tierID,
		BundleQuantity:  bundleQuantity,
		TotalEntries:    totalEntries,
		DailyAmount:     totalCost,
		Status:          "pending",
		AutoRenew:       true,
		NextBillingDate: time.Now().AddDate(0, 0, 1), // Tomorrow
		PaymentMethod:   paymentMethod,
		SubscriptionDate: time.Now(),
	}

	// Initialize payment
	paymentReq := PaymentRequest{
		Amount:      totalCost,
		Email:       "", // Will be filled from user profile
		Reference:   subscription.ID.String(),
		CallbackURL: "",
		Metadata: map[string]interface{}{
			"msisdn":          msisdn,
			"subscription_id": subscription.ID.String(),
			"type":            "subscription",
		},
	}
	paymentURL, err := s.paymentService.InitializePayment(ctx, paymentReq)
	if err != nil {
		return nil, "", fmt.Errorf("failed to initialize payment: %w", err)
	}
	paymentRef := subscription.ID.String() // Use subscription ID as reference

	subscription.PaymentReference = &paymentRef

	if err := s.tierRepo.CreateDailySubscription(ctx, subscription); err != nil {
		return nil, "", fmt.Errorf("failed to create subscription: %w", err)
	}

	return subscription, paymentURL, nil
}

// UpdateSubscriptionQuantity updates the bundle quantity for an existing subscription
func (s *SubscriptionTierService) UpdateSubscriptionQuantity(ctx context.Context, subscriptionID uuid.UUID, newBundleQuantity int) (*entities.DailySubscription, error) {
	subscription, err := s.tierRepo.FindDailySubscriptionByID(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("subscription not found: %w", err)
	}

	// Recalculate cost and entries
	totalCost, totalEntries, err := s.CalculateSubscriptionCost(ctx, subscription.TierID, newBundleQuantity)
	if err != nil {
		return nil, err
	}

	subscription.BundleQuantity = newBundleQuantity
	subscription.TotalEntries = totalEntries
	subscription.DailyAmount = totalCost

	if err := s.tierRepo.UpdateDailySubscription(ctx, subscription); err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	return subscription, nil
}

// CancelSubscription cancels a daily subscription
func (s *SubscriptionTierService) CancelSubscription(ctx context.Context, subscriptionID uuid.UUID, reason string) error {
	subscription, err := s.tierRepo.FindDailySubscriptionByID(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	now := time.Now()
	subscription.Status = "cancelled"
	subscription.AutoRenew = false
	subscription.CancelledAt = &now
	subscription.CancellationReason = reason

	if err := s.tierRepo.UpdateDailySubscription(ctx, subscription); err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	return nil
}

// PauseSubscription pauses a daily subscription
func (s *SubscriptionTierService) PauseSubscription(ctx context.Context, subscriptionID uuid.UUID) error {
	subscription, err := s.tierRepo.FindDailySubscriptionByID(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	subscription.Status = "paused"
	subscription.AutoRenew = false

	if err := s.tierRepo.UpdateDailySubscription(ctx, subscription); err != nil {
		return fmt.Errorf("failed to pause subscription: %w", err)
	}

	return nil
}

// ResumeSubscription resumes a paused subscription
func (s *SubscriptionTierService) ResumeSubscription(ctx context.Context, subscriptionID uuid.UUID) error {
	subscription, err := s.tierRepo.FindDailySubscriptionByID(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	subscription.Status = "active"
	subscription.AutoRenew = true
	subscription.NextBillingDate = time.Now().AddDate(0, 0, 1)

	if err := s.tierRepo.UpdateDailySubscription(ctx, subscription); err != nil {
		return fmt.Errorf("failed to resume subscription: %w", err)
	}

	return nil
}

// ProcessDailyBillings processes all subscriptions due for billing
func (s *SubscriptionTierService) ProcessDailyBillings(ctx context.Context) error {
	today := time.Now().Truncate(24 * time.Hour)
	subscriptions, err := s.tierRepo.FindDailySubscriptionsDueForBilling(ctx, today)
	if err != nil {
		return fmt.Errorf("failed to find subscriptions due for billing: %w", err)
	}

	for _, sub := range subscriptions {
		if err := s.processSingleBilling(ctx, sub); err != nil {
			// Log error but continue with other subscriptions
			fmt.Printf("Failed to process billing for subscription %s: %v\n", sub.ID, err)
		}
	}

	return nil
}

func (s *SubscriptionTierService) processSingleBilling(ctx context.Context, sub *entities.DailySubscription) error {
	// Create billing record
	billing := &entities.SubscriptionBilling{
		ID:             uuid.New(),
		SubscriptionID: sub.ID,
		MSISDN:         sub.MSISDN,
		BillingDate:    time.Now(),
		Amount:         sub.DailyAmount,
		EntriesAwarded: sub.TotalEntries,
		PointsEarned:   int(sub.DailyAmount / 20000), // ₦200 = 1 point (20000 kobo)
		Status:         "pending",
		PaymentMethod:  sub.PaymentMethod,
	}

	if err := s.tierRepo.CreateBilling(ctx, billing); err != nil {
		return fmt.Errorf("failed to create billing: %w", err)
	}

	// Process payment
	paymentReq := PaymentRequest{
		Amount:      sub.DailyAmount,
		Email:       "",
		Reference:   billing.ID.String(),
		CallbackURL: "",
		Metadata: map[string]interface{}{
			"msisdn":          sub.MSISDN,
			"subscription_id": sub.ID.String(),
			"billing_id":      billing.ID.String(),
			"type":            "subscription_billing",
		},
	}
	_, err := s.paymentService.InitializePayment(ctx, paymentReq)
	paymentRef := billing.ID.String()
	if err != nil {
		billing.Status = "failed"
		billing.FailureReason = err.Error()
		s.tierRepo.UpdateBilling(ctx, billing)
		return fmt.Errorf("payment failed: %w", err)
	}

	// Update billing as completed
	now := time.Now()
	billing.Status = "completed"
	billing.PaymentReference = paymentRef
	billing.ProcessedAt = &now

	if err := s.tierRepo.UpdateBilling(ctx, billing); err != nil {
		return fmt.Errorf("failed to update billing: %w", err)
	}

	// Update subscription
	sub.LastBillingDate = &now
	sub.NextBillingDate = now.AddDate(0, 0, 1)

	if err := s.tierRepo.UpdateDailySubscription(ctx, sub); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Update user points
	if sub.UserID != nil {
		user, err := s.userRepo.FindByID(ctx, *sub.UserID)
		if err == nil && user != nil {
			user.TotalPoints += billing.PointsEarned
			s.userRepo.Update(ctx, user)
		}
	}

	// Send notification
	// TODO: Fix SendSubscriptionRenewalNotification signature
	// s.notificationService.SendSubscriptionRenewalNotification(ctx, sub.MSISDN, sub.DailyAmount)

	return nil
}
