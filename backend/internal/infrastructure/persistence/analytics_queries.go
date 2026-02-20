package persistence

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// AnalyticsResult represents aggregate analytics data
type AnalyticsResult struct {
	TotalRevenue   float64
	TotalCount     int64
	AverageRevenue float64
}

// GetRevenueAnalytics performs optimized aggregate queries for revenue analytics
func GetRevenueAnalytics(ctx context.Context, db *gorm.DB) (map[string]interface{}, error) {
	// Calculate time boundaries
	now := time.Now()
	todayStart := now.Truncate(24 * time.Hour)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	
	// Total revenue and count (all time)
	var totalResult struct {
		TotalRevenue float64
		TotalCount   int64
	}
	err := db.WithContext(ctx).
		Table("transactions").
		Select("COALESCE(SUM(amount), 0) as total_revenue, COUNT(*) as total_count").
		Where("status = ? AND type = ?", "COMPLETED", "RECHARGE").
		Scan(&totalResult).Error
	if err != nil {
		return nil, err
	}
	
	// Today's revenue and count
	var todayResult struct {
		TotalRevenue float64
		TotalCount   int64
	}
	err = db.WithContext(ctx).
		Table("transactions").
		Select("COALESCE(SUM(amount), 0) as total_revenue, COUNT(*) as total_count").
		Where("status = ? AND type = ? AND created_at >= ?", "COMPLETED", "RECHARGE", todayStart).
		Scan(&todayResult).Error
	if err != nil {
		return nil, err
	}
	
	// This month's revenue and count
	var monthResult struct {
		TotalRevenue float64
		TotalCount   int64
	}
	err = db.WithContext(ctx).
		Table("transactions").
		Select("COALESCE(SUM(amount), 0) as total_revenue, COUNT(*) as total_count").
		Where("status = ? AND type = ? AND created_at >= ?", "COMPLETED", "RECHARGE", monthStart).
		Scan(&monthResult).Error
	if err != nil {
		return nil, err
	}
	
	// Calculate average
	var averageRecharge float64
	if totalResult.TotalCount > 0 {
		averageRecharge = totalResult.TotalRevenue / float64(totalResult.TotalCount)
	}
	
	return map[string]interface{}{
		"total_revenue":    totalResult.TotalRevenue,
		"total_recharges":  totalResult.TotalCount,
		"today_revenue":    todayResult.TotalRevenue,
		"today_recharges":  todayResult.TotalCount,
		"month_revenue":    monthResult.TotalRevenue,
		"month_recharges":  monthResult.TotalCount,
		"average_recharge": averageRecharge,
	}, nil
}

// GetUserAnalytics performs optimized aggregate queries for user analytics
func GetUserAnalytics(ctx context.Context, db *gorm.DB) (map[string]interface{}, error) {
	// Calculate time boundaries
	now := time.Now()
	todayStart := now.Truncate(24 * time.Hour)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	
	// Total users
	var totalUsers int64
	err := db.WithContext(ctx).
		Table("users").
		Where("deleted_at IS NULL").
		Count(&totalUsers).Error
	if err != nil {
		return nil, err
	}
	
	// Active users (users with at least one transaction)
	var activeUsers int64
	err = db.WithContext(ctx).
		Table("users").
		Joins("INNER JOIN transactions ON users.msisdn = transactions.msisdn").
		Where("users.deleted_at IS NULL").
		Group("users.id").
		Count(&activeUsers).Error
	if err != nil {
		return nil, err
	}
	
	// New users today
	var todayUsers int64
	err = db.WithContext(ctx).
		Table("users").
		Where("created_at >= ? AND deleted_at IS NULL", todayStart).
		Count(&todayUsers).Error
	if err != nil {
		return nil, err
	}
	
	// New users this month
	var monthUsers int64
	err = db.WithContext(ctx).
		Table("users").
		Where("created_at >= ? AND deleted_at IS NULL", monthStart).
		Count(&monthUsers).Error
	if err != nil {
		return nil, err
	}
	
	// Total points distributed
	var totalPoints int64
	err = db.WithContext(ctx).
		Table("users").
		Select("COALESCE(SUM(total_points), 0)").
		Where("deleted_at IS NULL").
		Scan(&totalPoints).Error
	if err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"total_users":   totalUsers,
		"active_users":  activeUsers,
		"today_users":   todayUsers,
		"month_users":   monthUsers,
		"total_points":  totalPoints,
	}, nil
}
