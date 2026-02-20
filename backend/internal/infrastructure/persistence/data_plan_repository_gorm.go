package persistence

import (
	"rechargemax/internal/domain/repositories"

	"gorm.io/gorm"
)

// NewDataPlanRepository creates a new GORM-based data plan repository
func NewDataPlanRepository(db *gorm.DB) repositories.DataPlanRepository {
	return repositories.NewDataPlanRepository(db)
}
