package persistence

import (
	"context"

	"golang.org/x/crypto/bcrypt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

// OTPRepositoryGORM implements OTPRepository using GORM
type OTPRepositoryGORM struct {
	db *gorm.DB
}

// NewOTPRepository creates a new OTP repository
func NewOTPRepository(db *gorm.DB) repositories.OTPRepository {
	return &OTPRepositoryGORM{db: db}
}

// Create creates a new OTP
func (r *OTPRepositoryGORM) Create(ctx context.Context, otp *entities.OTP) error {
	return r.db.WithContext(ctx).Create(otp).Error
}

// FindByID finds an OTP by ID
func (r *OTPRepositoryGORM) FindByID(ctx context.Context, id uuid.UUID) (*entities.OTP, error) {
	var otp entities.OTP
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&otp).Error
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

// FindValidOTP finds a valid (not used, not expired) OTP for a given MSISDN and code.
// OTPs are stored as bcrypt hashes; we fetch all unexpired records and verify each.
func (r *OTPRepositoryGORM) FindValidOTP(ctx context.Context, msisdn, code string) (*entities.OTP, error) {
	var otps []entities.OTP
	err := r.db.WithContext(ctx).
		Where("msisdn = ? AND is_used = ? AND expires_at > ?", msisdn, false, time.Now()).
		Order("created_at DESC").
		Limit(10).
		Find(&otps).Error
	if err != nil {
		return nil, err
	}
	for i := range otps {
		if bcrypt.CompareHashAndPassword([]byte(otps[i].Code), []byte(code)) == nil {
			return &otps[i], nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

// FindValidOTPWithPurpose finds a valid OTP with matching purpose.
// OTPs are stored as bcrypt hashes; we fetch recent unexpired records and verify each.
func (r *OTPRepositoryGORM) FindValidOTPWithPurpose(ctx context.Context, msisdn, code, purpose string) (*entities.OTP, error) {
	var otps []entities.OTP
	err := r.db.WithContext(ctx).
		Where("msisdn = ? AND purpose = ? AND is_used = ? AND expires_at > ?",
			msisdn, purpose, false, time.Now()).
		Order("created_at DESC").
		Limit(10).
		Find(&otps).Error
	if err != nil {
		return nil, err
	}
	for i := range otps {
		if bcrypt.CompareHashAndPassword([]byte(otps[i].Code), []byte(code)) == nil {
			return &otps[i], nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

// FindRecentOTPs finds OTPs created after a given time for rate limiting
func (r *OTPRepositoryGORM) FindRecentOTPs(ctx context.Context, msisdn string, since time.Time) ([]*entities.OTP, error) {
	var otps []*entities.OTP
	err := r.db.WithContext(ctx).
		Where("msisdn = ? AND created_at > ?", msisdn, since).
		Order("created_at DESC").
		Find(&otps).Error
	return otps, err
}

// CountRecentOTPs counts OTPs created after a given time for rate limiting
func (r *OTPRepositoryGORM) CountRecentOTPs(ctx context.Context, msisdn string, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.OTP{}).
		Where("msisdn = ? AND created_at > ?", msisdn, since).
		Count(&count).Error
	return count, err
}

// Update updates an OTP
func (r *OTPRepositoryGORM) Update(ctx context.Context, otp *entities.OTP) error {
	return r.db.WithContext(ctx).Save(otp).Error
}

// MarkAsUsed marks an OTP as used
func (r *OTPRepositoryGORM) MarkAsUsed(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entities.OTP{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_used":    true,
			"used_at":    now,
			"updated_at": now,
		}).Error
}

// DeleteExpired deletes expired OTPs (cleanup)
func (r *OTPRepositoryGORM) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&entities.OTP{}).Error
}

// DeleteOld deletes old used OTPs (cleanup)
func (r *OTPRepositoryGORM) DeleteOld(ctx context.Context, olderThan time.Time) error {
	return r.db.WithContext(ctx).
		Where("is_used = ? AND created_at < ?", true, olderThan).
		Delete(&entities.OTP{}).Error
}
