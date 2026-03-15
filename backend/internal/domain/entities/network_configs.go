package entities

import (
	"time"

)

// NetworkConfig represents the network_configs table
type NetworkConfig struct {
	ID             string     `json:"id" gorm:"column:id;type:uuid;primaryKey;default:uuid_generate_v4()"` 
	NetworkName    string     `json:"network_name" gorm:"column:network_name;not null" validate:"required"`
	NetworkCode    string     `json:"network_code" gorm:"column:network_code;uniqueIndex;not null" validate:"required"`
	IsActive       *bool      `json:"is_active" gorm:"column:is_active"`
	AirtimeEnabled *bool      `json:"airtime_enabled" gorm:"column:airtime_enabled"`
	DataEnabled    *bool      `json:"data_enabled" gorm:"column:data_enabled"`
	CommissionRate *float64   `json:"commission_rate" gorm:"column:commission_rate"`
	MinimumAmount  *int64     `json:"minimum_amount" gorm:"column:minimum_amount;type:bigint"`
	MaximumAmount  *int64     `json:"maximum_amount" gorm:"column:maximum_amount;type:bigint"`
	LogoUrl        string     `json:"logo_url" gorm:"column:logo_url"`
	BrandColor     string     `json:"brand_color" gorm:"column:brand_color"`
	SortOrder      *int       `json:"sort_order" gorm:"column:sort_order"`
	CreatedAt      *time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt      *time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// TableName specifies the table name for NetworkConfig
func (NetworkConfig) TableName() string {
	return "network_configs"
}
