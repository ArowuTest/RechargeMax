package persistence

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"rechargemax/internal/domain/entities"
	"rechargemax/internal/domain/repositories"
)

// NotificationRepositoryGORM implements the NotificationRepository interface using GORM
type NotificationRepositoryGORM struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new instance of NotificationRepositoryGORM
func NewNotificationRepository(db *gorm.DB) repositories.NotificationRepository {
	return &NotificationRepositoryGORM{db: db}
}

func (r *NotificationRepositoryGORM) Create(ctx context.Context, entity *entities.Notification) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

func (r *NotificationRepositoryGORM) FindByID(ctx context.Context, id uuid.UUID) (*entities.Notification, error) {
	var entity entities.Notification
	err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *NotificationRepositoryGORM) FindAll(ctx context.Context, limit, offset int) ([]*entities.Notification, error) {
	var notifications []*entities.Notification
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

func (r *NotificationRepositoryGORM) Update(ctx context.Context, entity *entities.Notification) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

func (r *NotificationRepositoryGORM) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.Notification{}, "id = ?", id).Error
}

func (r *NotificationRepositoryGORM) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.Notification{}).Count(&count).Error
	return count, err
}

// FindByMSISDN retrieves notifications for a specific user
func (r *NotificationRepositoryGORM) FindByMSISDN(ctx context.Context, msisdn string, limit, offset int) ([]*entities.Notification, error) {
	var notifications []*entities.Notification
	err := r.db.WithContext(ctx).
		Where("msisdn = ?", msisdn).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

// FindUnreadByMSISDN retrieves unread notifications for a user
func (r *NotificationRepositoryGORM) FindUnreadByMSISDN(ctx context.Context, msisdn string, limit, offset int) ([]*entities.Notification, error) {
	var notifications []*entities.Notification
	err := r.db.WithContext(ctx).
		Where("msisdn = ? AND is_read = ?", msisdn, false).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

// CountUnreadByMSISDN counts unread notifications for a user
func (r *NotificationRepositoryGORM) CountUnreadByMSISDN(ctx context.Context, msisdn string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.Notification{}).
		Where("msisdn = ? AND is_read = ?", msisdn, false).
		Count(&count).Error
	return count, err
}

// MarkAsRead marks a notification as read
func (r *NotificationRepositoryGORM) MarkAsRead(ctx context.Context, notificationID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entities.Notification{}).
		Where("id = ?", notificationID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).
		Error
}

// MarkAllAsRead marks all notifications as read for a user
func (r *NotificationRepositoryGORM) MarkAllAsRead(ctx context.Context, msisdn string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entities.Notification{}).
		Where("msisdn = ? AND is_read = ?", msisdn, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).
		Error
}

// FindByType retrieves notifications by type
func (r *NotificationRepositoryGORM) FindByType(ctx context.Context, notifType string, limit, offset int) ([]*entities.Notification, error) {
	var notifications []*entities.Notification
	err := r.db.WithContext(ctx).
		Where("type = ?", notifType).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

// FindByReference retrieves notifications by reference
func (r *NotificationRepositoryGORM) FindByReference(ctx context.Context, referenceType, referenceID string) ([]*entities.Notification, error) {
	var notifications []*entities.Notification
	err := r.db.WithContext(ctx).
		Where("reference_type = ? AND reference_id = ?", referenceType, referenceID).
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, err
}

// DeleteOldNotifications deletes notifications older than specified days
func (r *NotificationRepositoryGORM) DeleteOldNotifications(ctx context.Context, daysOld int) (int64, error) {
	cutoffDate := time.Now().AddDate(0, 0, -daysOld)
	result := r.db.WithContext(ctx).
		Where("created_at < ?", cutoffDate).
		Delete(&entities.Notification{})
	return result.RowsAffected, result.Error
}
