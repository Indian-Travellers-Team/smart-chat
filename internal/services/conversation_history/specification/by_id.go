package specification

import (
	"gorm.io/gorm"
)

// ByID filters conversations by their ID.
type ByID struct {
	ID uint
}

func (spec ByID) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("id = ?", spec.ID)
}
