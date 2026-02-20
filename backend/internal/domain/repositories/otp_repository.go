package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"

	"rechargemax/internal/domain/entities"
)

// OTPRepository defines the interface for OTP data access
type OTPRepository interface {
	// Create creates a new OTP
	Create(ctx context.Context, otp *entities.OTP) error

	// FindByID finds an OTP by ID
	FindByID(ctx context.Context, id uuid.UUID) (*entities.OTP, error)

	// FindValidOTP finds a valid (not used, not expired) OTP for a given MSISDN and code
	FindValidOTP(ctx context.Context, msisdn, code string) (*entities.OTP, error)

	// FindRecentOTPs finds OTPs created after a given time for rate limiting
	FindRecentOTPs(ctx context.Context, msisdn string, since time.Time) ([]*entities.OTP, error)

	// CountRecentOTPs counts OTPs created after a given time for rate limiting
	CountRecentOTPs(ctx context.Context, msisdn string, since time.Time) (int64, error)

	// Update updates an OTP
	Update(ctx context.Context, otp *entities.OTP) error

	// MarkAsUsed marks an OTP as used
	MarkAsUsed(ctx context.Context, id uuid.UUID) error

	// DeleteExpired deletes expired OTPs (cleanup)
	DeleteExpired(ctx context.Context) error

	// DeleteOld deletes old used OTPs (cleanup)
	DeleteOld(ctx context.Context, olderThan time.Time) error
}
