package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// NetworkCache represents the network_cache table
type NetworkCache struct {
	ID                   uuid.UUID      `json:"id" gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	MSISDN               string         `json:"msisdn" gorm:"column:msisdn;uniqueIndex;not null" validate:"required"`
	Network              string         `json:"network" gorm:"column:network;not null;index" validate:"required,oneof=MTN Airtel Glo 9mobile"`
	LastVerified         time.Time      `json:"last_verified" gorm:"column:last_verified_at;not null" validate:"required"`
	CacheExpires         time.Time      `json:"cache_expires" gorm:"column:cache_expires_at;not null;index" validate:"required"`
	LookupSource         string         `json:"lookup_source" gorm:"column:lookup_source;not null" validate:"required,oneof=hlr_api user_selection prefix_fallback"`
	HLRProvider          *string        `json:"hlr_provider" gorm:"column:hlr_provider"`
	HLRResponse          datatypes.JSON `json:"hlr_response" gorm:"column:hlr_response;type:jsonb"`
	IsValid              bool           `json:"is_valid" gorm:"column:is_valid;default:true;not null;index"`
	InvalidatedAt        *time.Time     `json:"invalidated_at" gorm:"column:invalidated_at"`
	InvalidationReason   *string        `json:"invalidation_reason" gorm:"column:invalidation_reason"`
	CreatedAt            time.Time      `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt            time.Time      `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	// DeletedAt is not used in this table (no soft delete)
}

// TableName specifies the table name for NetworkCache
func (NetworkCache) TableName() string {
	return "network_cache"
}

// IsExpired checks if the cache has expired
func (nc *NetworkCache) IsExpired() bool {
	return time.Now().After(nc.CacheExpires) || !nc.IsValid
}
