package entities

import (
	"time"
	"github.com/google/uuid"
)

// ServicePrice represents the service_pricing table
type ServicePrice struct {
	Id                             uuid.UUID                 `json:"id" gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	NetworkProvider                string                    `json:"network_provider" gorm:"column:network_provider;not null"`
	ServiceType                    string                    `json:"service_type" gorm:"column:service_type;not null"`
	DataBundleCode                 string                    `json:"data_bundle_code" gorm:"column:data_bundle_code"`
	BasePrice                      *int64                    `json:"base_price" gorm:"column:base_price;type:bigint;default:0"`
	SellingPrice                   *int64                    `json:"selling_price" gorm:"column:selling_price;type:bigint;not null"`
	CommissionRate                 *float64                  `json:"commission_rate" gorm:"column:commission_rate;type:decimal(5,2);default:0"`
	PlatformFee                    *float64                  `json:"platform_fee" gorm:"column:platform_fee;type:decimal(12,2);default:0"`
	MinAmount                      *int64                    `json:"min_amount" gorm:"column:min_amount;type:bigint;default:5000"`
	MaxAmount                      *int64                    `json:"max_amount" gorm:"column:max_amount;type:bigint;default:5000000"`
	IsActive                       *bool                     `json:"is_active" gorm:"column:is_active;default:true"`
	IsFeatured                     *bool                     `json:"is_featured" gorm:"column:is_featured;default:false"`
	SortOrder                      *int                      `json:"sort_order" gorm:"column:sort_order;default:0"`
	Description                    string                    `json:"description" gorm:"column:description"`
	ValidityPeriod                 string                    `json:"validity_period" gorm:"column:validity_period"`
	DataVolume                     string                    `json:"data_volume" gorm:"column:data_volume"`
	CreatedAt                      time.Time                 `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt                      *time.Time                `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName specifies the table name for ServicePrice
func (ServicePrice) TableName() string {
	return "service_pricing"
}
