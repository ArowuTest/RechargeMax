package persistence

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

// NetworkCacheRepositoryGORM implements the NetworkCacheRepository interface using GORM
type NetworkCacheRepositoryGORM struct {
	db *gorm.DB
}

// NewNetworkCacheRepository creates a new instance of NetworkCacheRepositoryGORM
func NewNetworkCacheRepository(db *gorm.DB) repositories.NetworkCacheRepository {
	return &NetworkCacheRepositoryGORM{db: db}
}

func (r *NetworkCacheRepositoryGORM) Create(ctx context.Context, entity *entities.NetworkCache) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

func (r *NetworkCacheRepositoryGORM) FindByID(ctx context.Context, id uuid.UUID) (*entities.NetworkCache, error) {
	var entity entities.NetworkCache
	err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *NetworkCacheRepositoryGORM) FindAll(ctx context.Context, limit, offset int) ([]*entities.NetworkCache, error) {
	var caches []*entities.NetworkCache
	err := r.db.WithContext(ctx).
		Order("last_verified_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&caches).Error
	return caches, err
}

func (r *NetworkCacheRepositoryGORM) Update(ctx context.Context, entity *entities.NetworkCache) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

func (r *NetworkCacheRepositoryGORM) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.NetworkCache{}, "id = ?", id).Error
}

func (r *NetworkCacheRepositoryGORM) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.NetworkCache{}).Count(&count).Error
	return count, err
}

// FindByMSISDN retrieves cache entry for a specific MSISDN
func (r *NetworkCacheRepositoryGORM) FindByMSISDN(ctx context.Context, msisdn string) (*entities.NetworkCache, error) {
	var cache entities.NetworkCache
	err := r.db.WithContext(ctx).
		Where("msisdn = ?", msisdn).
		First(&cache).Error
	if err != nil {
		return nil, err
	}
	return &cache, nil
}

// FindValidCache retrieves a valid (non-expired) cache entry
func (r *NetworkCacheRepositoryGORM) FindValidCache(ctx context.Context, msisdn string) (*entities.NetworkCache, error) {
	var cache entities.NetworkCache
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("msisdn = ? AND cache_expires_at > ? AND is_valid = ?", msisdn, now, true).
		First(&cache).Error
	if err != nil {
		return nil, err
	}
	return &cache, nil
}

// Invalidate marks a cache entry as invalid
func (r *NetworkCacheRepositoryGORM) Invalidate(ctx context.Context, msisdn string, reason string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entities.NetworkCache{}).
		Where("msisdn = ?", msisdn).
		Updates(map[string]interface{}{
			"is_valid":            false,
			"invalidated_at":      now,
			"invalidation_reason": reason,
		}).
		Error
}

// DeleteExpired deletes expired cache entries
func (r *NetworkCacheRepositoryGORM) DeleteExpired(ctx context.Context) (int64, error) {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Where("cache_expires_at < ?", now).
		Delete(&entities.NetworkCache{})
	return result.RowsAffected, result.Error
}

// CountByNetwork counts cache entries by network
func (r *NetworkCacheRepositoryGORM) CountByNetwork(ctx context.Context, network string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.NetworkCache{}).
		Where("network = ? AND is_valid = ?", network, true).
		Count(&count).Error
	return count, err
}

// CountByLookupSource counts cache entries by lookup source
func (r *NetworkCacheRepositoryGORM) CountByLookupSource(ctx context.Context, source string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.NetworkCache{}).
		Where("lookup_source = ?", source).
		Count(&count).Error
	return count, err
}
