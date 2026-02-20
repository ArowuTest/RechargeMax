package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"rechargemax/internal/domain/entities"
)

// USSDRechargeRepository defines operations for USSD recharges
type USSDRechargeRepository interface {
	// USSD Recharge Management
	Create(ctx context.Context, recharge *entities.USSDRecharge) error
	Update(ctx context.Context, recharge *entities.USSDRecharge) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.USSDRecharge, error)
	FindByTransactionRef(ctx context.Context, transactionRef string) (*entities.USSDRecharge, error)
	FindByMSISDN(ctx context.Context, msisdn string, startDate, endDate time.Time) ([]*entities.USSDRecharge, error)
	FindUnprocessed(ctx context.Context) ([]*entities.USSDRecharge, error)
	
	// Webhook Log Management
	CreateWebhookLog(ctx context.Context, log *entities.USSDWebhookLog) error
	UpdateWebhookLog(ctx context.Context, log *entities.USSDWebhookLog) error
	FindWebhookLogByID(ctx context.Context, id uuid.UUID) (*entities.USSDWebhookLog, error)
	FindWebhookLogs(ctx context.Context, provider string, startDate, endDate time.Time) ([]*entities.USSDWebhookLog, error)
	FindFailedWebhookLogs(ctx context.Context) ([]*entities.USSDWebhookLog, error)
}
