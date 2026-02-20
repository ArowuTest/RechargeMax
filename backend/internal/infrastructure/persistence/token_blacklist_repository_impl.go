package persistence

import (
	"context"
	"time"

	"gorm.io/gorm"

	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

type tokenBlacklistRepositoryImpl struct {
	db *gorm.DB
}

// NewTokenBlacklistRepository creates a new token blacklist repository
func NewTokenBlacklistRepository(db *gorm.DB) repositories.TokenBlacklistRepository {
	return &tokenBlacklistRepositoryImpl{db: db}
}

// Create adds a token to the blacklist
func (r *tokenBlacklistRepositoryImpl) Create(ctx context.Context, blacklist *entities.TokenBlacklist) error {
	return r.db.WithContext(ctx).Create(blacklist).Error
}

// IsBlacklisted checks if a token is blacklisted and not expired
func (r *tokenBlacklistRepositoryImpl) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.TokenBlacklist{}).
		Where("token = ? AND expires_at > ?", token, time.Now()).
		Count(&count).Error
	
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

// DeleteExpired removes expired tokens from the blacklist
func (r *tokenBlacklistRepositoryImpl) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at <= ?", time.Now()).
		Delete(&entities.TokenBlacklist{}).Error
}

// DeleteByAdminID removes all blacklisted tokens for an admin
func (r *tokenBlacklistRepositoryImpl) DeleteByAdminID(ctx context.Context, adminID string) error {
	return r.db.WithContext(ctx).
		Where("admin_id = ?", adminID).
		Delete(&entities.TokenBlacklist{}).Error
}
