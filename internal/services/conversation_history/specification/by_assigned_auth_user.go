package specification

import "gorm.io/gorm"

// ByAssignedAuthUser filters conversations assigned to a specific auth_users.user_id.
type ByAssignedAuthUser struct {
	AuthUserID uint
}

func (spec ByAssignedAuthUser) Apply(db *gorm.DB) *gorm.DB {
	return db.
		Joins("JOIN auth_user_conversation ON auth_user_conversation.conversation_id = conversations.id").
		Where("auth_user_conversation.auth_user_id = ?", spec.AuthUserID)
}
