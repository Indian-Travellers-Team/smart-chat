package specification

import (
	"time"

	"gorm.io/gorm"
)

// ByDateRange filters conversations within a specific date range.
type ByDateRange struct {
	StartDate, EndDate time.Time
}

func (spec ByDateRange) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("created_at >= ? AND created_at <= ?", spec.StartDate, spec.EndDate)
}
