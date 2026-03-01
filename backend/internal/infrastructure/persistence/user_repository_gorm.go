package persistence

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	
	"github.com/google/uuid"
	"gorm.io/gorm"
	
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

type userRepositoryGORM struct {
	db *gorm.DB
}

// NewUserRepository creates a new GORM implementation of UserRepository
func NewUserRepository(db *gorm.DB) repositories.UserRepository {
	return &userRepositoryGORM{db: db}
}

// CreateUserWithDefaults creates a new user with all default values set
// This is the SINGLE SOURCE OF TRUTH for user creation across the entire platform
// Used by: AuthService (OTP login), RechargeService (guest recharge), etc.
func (r *userRepositoryGORM) CreateUserWithDefaults(ctx context.Context, msisdn string, referredBy *uuid.UUID) (*entities.Users, error) {
	// Generate unique referral code for this user
	referralCode, err := generateUniqueReferralCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate referral code: %w", err)
	}
	
	// Create user with all defaults
	user := &entities.Users{
		ID:           uuid.New(),
		MSISDN:       msisdn,
		ReferralCode: referralCode,
		ReferredBy:   referredBy,
		LoyaltyTier:  "BRONZE",
		TotalPoints:  0,
		IsActive:     true,
		IsVerified:   false,
		KYCStatus:    "PENDING",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	// Create in database
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, err
	}
	
	return user, nil
}

// generateUniqueReferralCode generates a unique referral code
// Format: REF + 8 random uppercase hex characters (e.g., REF1A2B3C4D)
func generateUniqueReferralCode() (string, error) {
	bytes := make([]byte, 4) // 4 bytes = 8 hex characters
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "REF" + strings.ToUpper(hex.EncodeToString(bytes)), nil
}

// Create creates a new user
func (r *userRepositoryGORM) Create(ctx context.Context, user *entities.Users) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// CreateBatch creates multiple users in a single transaction
func (r *userRepositoryGORM) CreateBatch(ctx context.Context, users []*entities.Users) error {
	return r.db.WithContext(ctx).CreateInBatches(users, 100).Error
}

// FindByID finds a user by ID
func (r *userRepositoryGORM) FindByID(ctx context.Context, id uuid.UUID) (*entities.Users, error) {
	var user entities.Users
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByMSISDN finds a user by phone number
func (r *userRepositoryGORM) FindByMSISDN(ctx context.Context, msisdn string) (*entities.Users, error) {
	var user entities.Users
	err := r.db.WithContext(ctx).Where("msisdn = ?", msisdn).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByReferralCode finds a user by referral code
func (r *userRepositoryGORM) FindByReferralCode(ctx context.Context, code string) (*entities.Users, error) {
	var user entities.Users
	err := r.db.WithContext(ctx).Where("referral_code = ?", code).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail finds a user by email
func (r *userRepositoryGORM) FindByEmail(ctx context.Context, email string) (*entities.Users, error) {
	var user entities.Users
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindAll retrieves all users with pagination
func (r *userRepositoryGORM) FindAll(ctx context.Context, limit, offset int) ([]*entities.Users, error) {
	var users []*entities.Users
	err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&users).Error
	return users, err
}

// FindByLoyaltyTier finds users by loyalty tier
func (r *userRepositoryGORM) FindByLoyaltyTier(ctx context.Context, tier string) ([]*entities.Users, error) {
	var users []*entities.Users
	err := r.db.WithContext(ctx).
		Where("loyalty_tier = ?", tier).
		Find(&users).Error
	return users, err
}

// FindActiveUsers retrieves active users
func (r *userRepositoryGORM) FindActiveUsers(ctx context.Context, limit, offset int) ([]*entities.Users, error) {
	var users []*entities.Users
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Limit(limit).
		Offset(offset).
		Order("last_login_at DESC NULLS LAST").
		Find(&users).Error
	return users, err
}

// Count returns the total number of users
func (r *userRepositoryGORM) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.Users{}).Count(&count).Error
	return count, err
}

// Update updates a user
func (r *userRepositoryGORM) Update(ctx context.Context, user *entities.Users) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// UpdateStatus updates only the is_active field for a user (avoids full-row update and unique constraint issues)
func (r *userRepositoryGORM) UpdateStatus(ctx context.Context, userID uuid.UUID, isActive bool) error {
	return r.db.WithContext(ctx).
		Model(&entities.Users{}).
		Where("id = ?", userID).
		Update("is_active", isActive).
		Error
}

// UpdatePoints updates user points
func (r *userRepositoryGORM) UpdatePoints(ctx context.Context, userID uuid.UUID, points int) error {
	return r.db.WithContext(ctx).
		Model(&entities.Users{}).
		Where("id = ?", userID).
		Update("total_points", gorm.Expr("total_points + ?", points)).
		Error
}

// UpdateLoyaltyTier updates user loyalty tier
func (r *userRepositoryGORM) UpdateLoyaltyTier(ctx context.Context, userID uuid.UUID, tier string) error {
	return r.db.WithContext(ctx).
		Model(&entities.Users{}).
		Where("id = ?", userID).
		Update("loyalty_tier", tier).
		Error
}

// UpdateRechargeStats updates user recharge statistics
func (r *userRepositoryGORM) UpdateRechargeStats(ctx context.Context, userID uuid.UUID, amount float64) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entities.Users{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"total_recharge_amount": gorm.Expr("total_recharge_amount + ?", amount),
			"total_transactions":    gorm.Expr("total_transactions + 1"),
			"last_recharge_date":    now,
		}).
		Error
}

// IncrementReferrals increments user referral count
func (r *userRepositoryGORM) IncrementReferrals(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entities.Users{}).
		Where("id = ?", userID).
		Update("total_referrals", gorm.Expr("total_referrals + 1")).
		Error
}

// UpdateLastLogin updates user last login timestamp
func (r *userRepositoryGORM) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entities.Users{}).
		Where("id = ?", userID).
		Update("last_login_at", now).
		Error
}

// Delete permanently deletes a user
func (r *userRepositoryGORM) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Delete(&entities.Users{}, "id = ?", id).
		Error
}

// SoftDelete soft deletes a user (sets is_active to false)
func (r *userRepositoryGORM) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entities.Users{}).
		Where("id = ?", id).
		Update("is_active", false).
		Error
}

// GetTopReferrers gets users with most referrals
func (r *userRepositoryGORM) GetTopReferrers(ctx context.Context, limit int) ([]*entities.Users, error) {
	var users []*entities.Users
	err := r.db.WithContext(ctx).
		Where("total_referrals > 0").
		Order("total_referrals DESC").
		Limit(limit).
		Find(&users).Error
	return users, err
}

// GetUsersByDateRange gets users created within a date range
func (r *userRepositoryGORM) GetUsersByDateRange(ctx context.Context, startDate, endDate string) ([]*entities.Users, error) {
	var users []*entities.Users
	err := r.db.WithContext(ctx).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate).
		Order("created_at DESC").
		Find(&users).Error
	return users, err
}

// GetUserStatistics returns user statistics
func (r *userRepositoryGORM) GetUserStatistics(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Total users
	var totalUsers int64
	if err := r.db.WithContext(ctx).Model(&entities.Users{}).Count(&totalUsers).Error; err != nil {
		return nil, err
	}
	stats["total_users"] = totalUsers
	
	// Active users
	var activeUsers int64
	if err := r.db.WithContext(ctx).Model(&entities.Users{}).Where("is_active = ?", true).Count(&activeUsers).Error; err != nil {
		return nil, err
	}
	stats["active_users"] = activeUsers
	
	// Verified users
	var verifiedUsers int64
	if err := r.db.WithContext(ctx).Model(&entities.Users{}).Where("is_verified = ?", true).Count(&verifiedUsers).Error; err != nil {
		return nil, err
	}
	stats["verified_users"] = verifiedUsers
	
	// Users by loyalty tier
	var tierCounts []struct {
		LoyaltyTier string
		Count       int64
	}
	if err := r.db.WithContext(ctx).
		Model(&entities.Users{}).
		Select("loyalty_tier, COUNT(*) as count").
		Group("loyalty_tier").
		Scan(&tierCounts).Error; err != nil {
		return nil, err
	}
	stats["users_by_tier"] = tierCounts
	
	return stats, nil
}


// CountByDateRange counts users created within a date range (for analytics)
func (r *userRepositoryGORM) CountByDateRange(ctx context.Context, startDate, endDate string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.Users{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&count).Error
	return count, err
}
