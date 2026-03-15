package entities

import (
	"time"

	"gorm.io/datatypes"
)

// AdminUser represents the admin_users table
type AdminUser struct {
	ID            string         `json:"id" gorm:"column:id;primaryKey"`
	Email         string         `json:"email" gorm:"column:email;uniqueIndex;not null" validate:"required,email"`
	PasswordHash  string         `json:"password_hash" gorm:"column:password_hash;not null" validate:"required"`
	FullName      string         `json:"full_name" gorm:"column:full_name;not null" validate:"required"`
	Role          string         `json:"role" gorm:"column:role;not null"`
	Permissions   datatypes.JSON `json:"permissions" gorm:"column:permissions"`
	IsActive      *bool          `json:"is_active" gorm:"column:is_active"`
	LastLogin     *time.Time     `json:"last_login" gorm:"column:last_login"`
	LastLoginAt   *time.Time     `json:"last_login_at" gorm:"column:last_login_at"`
	LoginAttempts *int           `json:"login_attempts" gorm:"column:login_attempts"`
	LockedUntil   *time.Time     `json:"locked_until" gorm:"column:locked_until"`
	CreatedAt     *time.Time     `json:"created_at" gorm:"column:created_at"`
	UpdatedAt     *time.Time     `json:"updated_at" gorm:"column:updated_at"`
}

// TableName specifies the table name for AdminUser
func (AdminUser) TableName() string {
	return "admin_users"
}
