package models

import "gorm.io/gorm"

type AuthUserConversation struct {
	gorm.Model
	AuthUserID     uint         `gorm:"column:auth_user_id;not null;index:idx_auth_user_conversation,unique"`
	AuthUser       AuthUser     `gorm:"foreignKey:AuthUserID;references:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ConversationID uint         `gorm:"column:conversation_id;not null;index:idx_auth_user_conversation,unique"`
	Conversation   Conversation `gorm:"foreignKey:ConversationID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Started        bool         `gorm:"column:started;not null;default:false"`
	Resolved       bool         `gorm:"column:resolved;not null;default:false"`
	Comments       string       `gorm:"column:comments;type:text;not null;default:''"`
}

func (AuthUserConversation) TableName() string {
	return "auth_user_conversation"
}
