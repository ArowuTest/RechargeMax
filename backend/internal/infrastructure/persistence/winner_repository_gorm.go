package persistence

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

// WinnerRepositoryGORM implements the WinnerRepository interface using GORM
type WinnerRepositoryGORM struct {
	db *gorm.DB
}

// NewWinnerRepository creates a new instance of WinnerRepositoryGORM
func NewWinnerRepository(db *gorm.DB) repositories.WinnerRepository {
	return &WinnerRepositoryGORM{db: db}
}

func (r *WinnerRepositoryGORM) Create(ctx context.Context, entity *entities.Winner) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

func (r *WinnerRepositoryGORM) FindByID(ctx context.Context, id uuid.UUID) (*entities.Winner, error) {
	var entity entities.Winner
	err := r.db.WithContext(ctx).
		Preload("Draw").
		First(&entity, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *WinnerRepositoryGORM) FindAll(ctx context.Context, limit, offset int) ([]*entities.Winner, error) {
	var winners []*entities.Winner
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&winners).Error
	return winners, err
}

func (r *WinnerRepositoryGORM) Update(ctx context.Context, entity *entities.Winner) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

func (r *WinnerRepositoryGORM) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.Winner{}, "id = ?", id).Error
}

func (r *WinnerRepositoryGORM) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.Winner{}).Count(&count).Error
	return count, err
}

// FindByDrawID retrieves all winners for a specific draw
func (r *WinnerRepositoryGORM) FindByDrawID(ctx context.Context, drawID uuid.UUID) ([]*entities.Winner, error) {
	var winners []*entities.Winner
	err := r.db.WithContext(ctx).
		Where("draw_id = ?", drawID).
		Order("position ASC").
		Find(&winners).Error
	return winners, err
}

// FindByMSISDN retrieves all wins for a specific user
func (r *WinnerRepositoryGORM) FindByMSISDN(ctx context.Context, msisdn string, limit, offset int) ([]*entities.Winner, error) {
	var winners []*entities.Winner
	err := r.db.WithContext(ctx).
		Where("msisdn = ?", msisdn).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&winners).Error
	return winners, err
}

// FindByClaimStatus retrieves winners by claim status
func (r *WinnerRepositoryGORM) FindByClaimStatus(ctx context.Context, status string, limit, offset int) ([]*entities.Winner, error) {
	var winners []*entities.Winner
	err := r.db.WithContext(ctx).
		Where("claim_status = ?", status).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&winners).Error
	return winners, err
}

// FindByPayoutStatus retrieves winners by payout status
func (r *WinnerRepositoryGORM) FindByPayoutStatus(ctx context.Context, status string, limit, offset int) ([]*entities.Winner, error) {
	var winners []*entities.Winner
	err := r.db.WithContext(ctx).
		Where("payout_status = ?", status).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&winners).Error
	return winners, err
}

// FindByProvisionStatus retrieves winners by provision status
func (r *WinnerRepositoryGORM) FindByProvisionStatus(ctx context.Context, status string, limit, offset int) ([]*entities.Winner, error) {
	var winners []*entities.Winner
	err := r.db.WithContext(ctx).
		Where("provision_status = ?", status).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&winners).Error
	return winners, err
}

// FindPendingProvisioning retrieves winners with auto_provision=true and provision_status=pending
func (r *WinnerRepositoryGORM) FindPendingProvisioning(ctx context.Context) ([]*entities.Winner, error) {
	var winners []*entities.Winner
	err := r.db.WithContext(ctx).
		Where("auto_provision = ? AND provision_status = ?", true, "pending").
		Order("created_at ASC").
		Find(&winners).Error
	return winners, err
}

// FindExpiredClaims retrieves winners with expired claim deadlines
func (r *WinnerRepositoryGORM) FindExpiredClaims(ctx context.Context) ([]*entities.Winner, error) {
	var winners []*entities.Winner
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("claim_status = ? AND claim_deadline < ?", "pending", now).
		Order("claim_deadline ASC").
		Find(&winners).Error
	return winners, err
}

// CountByMSISDN counts total wins for a user
func (r *WinnerRepositoryGORM) CountByMSISDN(ctx context.Context, msisdn string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.Winner{}).
		Where("msisdn = ?", msisdn).
		Count(&count).Error
	return count, err
}

// UpdateClaimStatus updates the claim status of a winner
func (r *WinnerRepositoryGORM) UpdateClaimStatus(ctx context.Context, winnerID uuid.UUID, status string) error {
	updates := map[string]interface{}{
		"claim_status": status,
	}
	if status == "claimed" {
		now := time.Now()
		updates["claimed_at"] = now
	}
	return r.db.WithContext(ctx).
		Model(&entities.Winner{}).
		Where("id = ?", winnerID).
		Updates(updates).
		Error
}

// UpdatePayoutStatus updates the payout status of a winner
func (r *WinnerRepositoryGORM) UpdatePayoutStatus(ctx context.Context, winnerID uuid.UUID, status string) error {
	return r.db.WithContext(ctx).
		Model(&entities.Winner{}).
		Where("id = ?", winnerID).
		Update("payout_status", status).
		Error
}

// UpdateProvisionStatus updates the provision status of a winner
func (r *WinnerRepositoryGORM) UpdateProvisionStatus(ctx context.Context, winnerID uuid.UUID, status string) error {
	updates := map[string]interface{}{
		"provision_status": status,
	}
	if status == "provisioned" {
		now := time.Now()
		updates["provisioned_at"] = now
	}
	return r.db.WithContext(ctx).
		Model(&entities.Winner{}).
		Where("id = ?", winnerID).
		Updates(updates).
		Error
}


// FindUnclaimedByUserID finds all unclaimed prizes for a specific user
func (r *WinnerRepositoryGORM) FindUnclaimedByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Winner, error) {
	var winners []*entities.Winner
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND claim_status = ?", userID, "unclaimed").
		Order("created_at DESC").
		Find(&winners).Error
	return winners, err
}

// CountUnclaimedByUserID counts unclaimed prizes for a specific user
func (r *WinnerRepositoryGORM) CountUnclaimedByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.Winner{}).
		Where("user_id = ? AND claim_status = ?", userID, "unclaimed").
		Count(&count).Error
	return count, err
}

// FindUnclaimedBeforeDeadline finds all unclaimed prizes before a specific deadline
func (r *WinnerRepositoryGORM) FindUnclaimedBeforeDeadline(ctx context.Context, deadline string) ([]*entities.Winner, error) {
	var winners []*entities.Winner
	err := r.db.WithContext(ctx).
		Where("claim_status = ? AND claim_deadline < ?", "unclaimed", deadline).
		Order("claim_deadline ASC").
		Find(&winners).Error
	return winners, err
}
