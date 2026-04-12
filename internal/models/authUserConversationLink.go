package models

import "gorm.io/gorm"

type AuthUserConversation struct {
	gorm.Model
	AuthUserID     uint         `gorm:"column:auth_user_id;not null;index:idx_auth_user_conversation,unique"`
	AuthUser       AuthUser     `gorm:"foreignKey:AuthUserID;references:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ConversationID uint         `gorm:"column:conversation_id;not null;index:idx_auth_user_conversation,unique"`
	Conversation   Conversation `gorm:"foreignKey:ConversationID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (AuthUserConversation) TableName() string {
	return "auth_user_conversation"
}
