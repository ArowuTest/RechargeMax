package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"rechargemax/internal/domain/entities"
)

// SubscriptionTierRepository defines operations for subscription tiers
type SubscriptionTierRepository interface {
	// Tier Management
	Create(ctx context.Context, tier *entities.SubscriptionTier) error
	Update(ctx context.Context, tier *entities.SubscriptionTier) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.SubscriptionTier, error)
	FindAll(ctx context.Context) ([]*entities.SubscriptionTier, error)
	FindActive(ctx context.Context) ([]*entities.SubscriptionTier, error)

	// Pricing Management
	CreatePricing(ctx context.Context, pricing *entities.SubscriptionPricing) error
	UpdatePricing(ctx context.Context, pricing *entities.SubscriptionPricing) error
	GetCurrentPricing(ctx context.Context) (*entities.SubscriptionPricing, error)
	GetPricingHistory(ctx context.Context) ([]*entities.SubscriptionPricing, error)

	// Daily Subscription Management
	CreateDailySubscription(ctx context.Context, sub *entities.DailySubscription) error
	UpdateDailySubscription(ctx context.Context, sub *entities.DailySubscription) error
	FindDailySubscriptionByID(ctx context.Context, id uuid.UUID) (*entities.DailySubscription, error)
	FindDailySubscriptionsByMSISDN(ctx context.Context, msisdn string) ([]*entities.DailySubscription, error)
	FindActiveDailySubscriptions(ctx context.Context) ([]*entities.DailySubscription, error)
	FindDailySubscriptionsDueForBilling(ctx context.Context, date time.Time) ([]*entities.DailySubscription, error)

	// Billing Management
	CreateBilling(ctx context.Context, billing *entities.SubscriptionBilling) error
	UpdateBilling(ctx context.Context, billing *entities.SubscriptionBilling) error
	FindBillingByID(ctx context.Context, id uuid.UUID) (*entities.SubscriptionBilling, error)
	FindBillingsBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) ([]*entities.SubscriptionBilling, error)
	FindBillingsByMSISDN(ctx context.Context, msisdn string, startDate, endDate time.Time) ([]*entities.SubscriptionBilling, error)
	FindPendingBillings(ctx context.Context) ([]*entities.SubscriptionBilling, error)
}
