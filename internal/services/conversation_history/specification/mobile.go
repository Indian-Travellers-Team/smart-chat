package specification

import "gorm.io/gorm"

// ByMobile filters conversations based on the mobile number.
// It joins the Session and User associations so you can filter on the user's mobile number.
type ByMobile struct {
	Mobile string
}

func (spec ByMobile) Apply(db *gorm.DB) *gorm.DB {
	return db.Joins("Session").Joins("Session.User").Where("users.mobile = ?", spec.Mobile)
}
