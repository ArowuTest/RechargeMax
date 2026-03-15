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

	// FindValidOTPWithPurpose finds a valid OTP with matching purpose
	FindValidOTPWithPurpose(ctx context.Context, msisdn, code, purpose string) (*entities.OTP, error)

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

	// FindLatestPendingOTP returns the most recently created, unused, unexpired OTP
	// for the given msisdn+purpose (without code verification). Used to increment
	// the failed attempt counter when the submitted code was wrong.
	FindLatestPendingOTP(ctx context.Context, msisdn, purpose string) (*entities.OTP, error)

	// IncrementFailedAttempts increments the failed attempt counter for an OTP.
	// Returns the new count so the caller can decide whether to invalidate.
	IncrementFailedAttempts(ctx context.Context, id uuid.UUID) (int, error)

	// InvalidateByMSISDN marks all pending OTPs for an MSISDN as used so the
	// attacker must request a fresh OTP after too many failures (SEC-008).
	InvalidateByMSISDN(ctx context.Context, msisdn, purpose string) error
}
