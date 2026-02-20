package entities

import (
	"time"

	"gorm.io/datatypes"
)

// UserNotificationPreferences represents the user_notification_preferences table
type UserNotificationPreferences struct {
	TransactionNotifications datatypes.JSON `json:"transaction_notifications" gorm:"column:transaction_notifications"`
	PrizeNotifications       datatypes.JSON `json:"prize_notifications" gorm:"column:prize_notifications"`
	DrawNotifications        datatypes.JSON `json:"draw_notifications" gorm:"column:draw_notifications"`
	AffiliateNotifications   datatypes.JSON `json:"affiliate_notifications" gorm:"column:affiliate_notifications"`
	PromotionalNotifications datatypes.JSON `json:"promotional_notifications" gorm:"column:promotional_notifications"`
	SecurityNotifications    datatypes.JSON `json:"security_notifications" gorm:"column:security_notifications"`
	DoNotDisturbStart        string         `json:"do_not_disturb_start" gorm:"column:do_not_disturb_start"`
	DoNotDisturbEnd          string         `json:"do_not_disturb_end" gorm:"column:do_not_disturb_end"`
	Timezone                 string         `json:"timezone" gorm:"column:timezone"`
	PreferredLanguage        string         `json:"preferred_language" gorm:"column:preferred_language"`
	CreatedAt                *time.Time     `json:"created_at" gorm:"column:created_at"`
	UpdatedAt                *time.Time     `json:"updated_at" gorm:"column:updated_at"`
}

// TableName specifies the table name for UserNotificationPreferences
func (UserNotificationPreferences) TableName() string {
	return "user_notification_preferences"
}
