package persistence

import (
	"context"
	"time"
	
	"github.com/google/uuid"
	"gorm.io/gorm"
	
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

type transactionRepositoryGORM struct {
	db *gorm.DB
}

// NewTransactionRepository creates a new GORM implementation
func NewTransactionRepository(db *gorm.DB) repositories.TransactionRepository {
	return &transactionRepositoryGORM{db: db}
}

// Create creates a new record
func (r *transactionRepositoryGORM) Create(ctx context.Context, entity *entities.Transactions) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// FindByID finds a record by ID
func (r *transactionRepositoryGORM) FindByID(ctx context.Context, id uuid.UUID) (*entities.Transactions, error) {
	var entity entities.Transactions
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// FindAll retrieves all records with pagination
func (r *transactionRepositoryGORM) FindAll(ctx context.Context, limit, offset int) ([]*entities.Transactions, error) {
	var entities []*entities.Transactions
	err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&entities).Error
	return entities, err
}

// Update updates a record
func (r *transactionRepositoryGORM) Update(ctx context.Context, entity *entities.Transactions) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

// Delete deletes a record
func (r *transactionRepositoryGORM) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.Transactions{}, "id = ?", id).Error
}

// Count returns the total number of records
func (r *transactionRepositoryGORM) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.Transactions{}).Count(&count).Error
	return count, err
}

// FindByUserID finds transactions for a specific user
func (r *transactionRepositoryGORM) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Transactions, error) {
	var transactions []*entities.Transactions
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&transactions).Error
	return transactions, err
}

// FindByReference finds a transaction by reference
func (r *transactionRepositoryGORM) FindByReference(ctx context.Context, reference string) (*entities.Transactions, error) {
	var transaction entities.Transactions
	err := r.db.WithContext(ctx).
		Where("reference = ?", reference).
		First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// FindByStatus finds transactions by status
func (r *transactionRepositoryGORM) FindByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.Transactions, error) {
	var transactions []*entities.Transactions
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&transactions).Error
	return transactions, err
}

// GetTotalRevenue calculates total revenue
func (r *transactionRepositoryGORM) GetTotalRevenue(ctx context.Context) (float64, error) {
	var total float64
	err := r.db.WithContext(ctx).
		Model(&entities.Transactions{}).
		Where("status = ?", "completed").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}

// GetRevenueByDate calculates revenue for a specific date
func (r *transactionRepositoryGORM) GetRevenueByDate(ctx context.Context, date time.Time) (float64, error) {
	var total float64
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	
	err := r.db.WithContext(ctx).
		Model(&entities.Transactions{}).
		Where("status = ? AND created_at >= ? AND created_at < ?", "completed", startOfDay, endOfDay).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}

// CountPendingWithdrawals counts pending withdrawal transactions
func (r *transactionRepositoryGORM) CountPendingWithdrawals(ctx context.Context) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.Transactions{}).
		Where("type = ? AND status = ?", "withdrawal", "pending").
		Count(&count).Error
	return int(count), err
}

// CountActiveSubscriptions counts active subscription transactions
func (r *transactionRepositoryGORM) CountActiveSubscriptions(ctx context.Context) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.Transactions{}).
		Where("type = ? AND status = ?", "subscription", "active").
		Count(&count).Error
	return int(count), err
}

// CountSpinsByDate counts spins for a specific date
func (r *transactionRepositoryGORM) CountSpinsByDate(ctx context.Context, date time.Time) (int, error) {
	var count int64
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	
	err := r.db.WithContext(ctx).
		Model(&entities.Transactions{}).
		Where("type = ? AND created_at >= ? AND created_at < ?", "spin", startOfDay, endOfDay).
		Count(&count).Error
	return int(count), err
}

// CountEligibleForSpin counts transactions eligible for spin (amount >= 1000, no spin yet)
func (r *transactionRepositoryGORM) CountEligibleForSpin(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.Transactions{}).
		Where("user_id = ? AND amount >= ? AND status = ? AND spin_claimed = ?", 
			userID, 100000, "completed", false). // 100000 kobo = ₦1000
		Count(&count).Error
	return count, err
}

// FindEligibleForSpin finds a transaction eligible for spin
func (r *transactionRepositoryGORM) FindEligibleForSpin(ctx context.Context, userID uuid.UUID) (*entities.Transactions, error) {
	var transaction entities.Transactions
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND amount >= ? AND status = ? AND spin_claimed = ?", 
			userID, 100000, "completed", false). // 100000 kobo = ₦1000
		Order("created_at ASC").
		First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// CountByUserID counts SUCCESSFUL transactions for a specific user.
// Only successful recharges count toward the first-recharge commission gate —
// counting all statuses (pending/failed) would wrongly block commission on
// retried payments.
func (r *transactionRepositoryGORM) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.Transactions{}).
		Where("user_id = ? AND status = 'SUCCESS'", userID).
		Count(&count).Error
	return count, err
}

// FindByPaymentRef finds a transaction by payment reference
func (r *transactionRepositoryGORM) FindByPaymentRef(ctx context.Context, paymentRef string) (*entities.Transactions, error) {
	var transaction entities.Transactions
	err := r.db.WithContext(ctx).Where("payment_reference = ?", paymentRef).First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// FindByMSISDN finds transactions by MSISDN with pagination
func (r *transactionRepositoryGORM) FindByMSISDN(ctx context.Context, msisdn string, limit, offset int) ([]*entities.Transactions, error) {
	var transactions []*entities.Transactions
	err := r.db.WithContext(ctx).
		Where("msisdn = ?", msisdn).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error
	return transactions, err
}

// FindByPaymentReference finds a transaction by payment reference (for webhook processing)
func (r *transactionRepositoryGORM) FindByPaymentReference(ctx context.Context, reference string) (*entities.Transactions, error) {
	var entity entities.Transactions
	err := r.db.WithContext(ctx).
		Where("payment_reference = ?", reference).
		First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// CountByUserIDAndDateRange counts transactions for a user within a date range (for fraud detection)
func (r *transactionRepositoryGORM) CountByUserIDAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.Transactions{}).
		Where("user_id = ? AND created_at BETWEEN ? AND ?", userID, start, end).
		Count(&count).Error
	return count, err
}

// SumAmountByUserIDAndDateRange sums transaction amounts for a user within a date range (for fraud detection)
func (r *transactionRepositoryGORM) SumAmountByUserIDAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) (float64, error) {
	var result struct {
		Total float64
	}
	err := r.db.WithContext(ctx).
		Model(&entities.Transactions{}).
		Select("COALESCE(SUM(amount), 0) as total").
		Where("user_id = ? AND created_at BETWEEN ? AND ? AND status = ?", userID, start, end, "SUCCESS").
		Scan(&result).Error
	return result.Total, err
}
