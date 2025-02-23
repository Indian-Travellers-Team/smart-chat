package models

import (
	"gorm.io/gorm"
)

type FunctionCall struct {
	gorm.Model
	ConversationID   uint         `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Conversation     Conversation `gorm:"foreignKey:ConversationID;references:ID"`
	MessageID        uint         `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	MessagePair      MessagePair  `gorm:"foreignKey:MessageID;references:ID"`
	Name             string       `gorm:"type:varchar(100)"`
	Args             []byte       `gorm:"type:json"`
	FunctionResponse string       `gorm:"type:varchar(500)"`
}
