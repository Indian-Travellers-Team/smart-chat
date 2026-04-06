package specification

import (
	"strings"

	"gorm.io/gorm"
)

// BySource filters conversations based on session source.
type BySource struct {
	Source string
}

func (spec BySource) Apply(db *gorm.DB) *gorm.DB {
	source := strings.ToLower(strings.TrimSpace(spec.Source))
	return db.
		Joins("JOIN sessions ON sessions.id = conversations.session_id").
		Where("LOWER(sessions.source) = ?", source)
}
