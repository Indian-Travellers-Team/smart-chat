package specification

import "gorm.io/gorm"

// ByMobile filters conversations based on the mobile number.
type ByMobile struct {
	Mobile string
}

func (spec ByMobile) Apply(db *gorm.DB) *gorm.DB {
	return db.
		Joins("JOIN sessions ON sessions.id = conversations.session_id").
		Joins("JOIN users ON users.id = sessions.user_id").
		Where("users.mobile = ?", spec.Mobile)
}
