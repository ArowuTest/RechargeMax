package entities

import (
	"time"

	"gorm.io/datatypes"
)

// ApplicationMetric represents the application_metrics table
type ApplicationMetric struct {
	MetricName  string         `json:"metric_name" gorm:"column:metric_name;not null" validate:"required"`
	MetricValue float64        `json:"metric_value" gorm:"column:metric_value;not null" validate:"required"`
	MetricUnit  string         `json:"metric_unit" gorm:"column:metric_unit"`
	Tags        datatypes.JSON `json:"tags" gorm:"column:tags"`
	Dimensions  datatypes.JSON `json:"dimensions" gorm:"column:dimensions"`
	RecordedAt  *time.Time     `json:"recorded_at" gorm:"column:recorded_at"`
}

// TableName specifies the table name for ApplicationMetric
func (ApplicationMetric) TableName() string {
	return "application_metrics"
}
