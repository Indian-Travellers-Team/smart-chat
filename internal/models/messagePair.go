package models

import (
	"gorm.io/gorm"
)

type MessageType int8

const (
	MessageTypeUnknown MessageType = iota
	MessageTypeUserFix
	MessageTypeUserSent
	MessageTypeOffTopic
	MessageTypeFunctionCall
)

type MessagePair struct {
	gorm.Model
	ConversationID uint         `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Conversation   Conversation `gorm:"foreignKey:ConversationID;references:ID"`
	User           string       `gorm:"type:text"`
	Bot            string       `gorm:"type:text"`
	BotSummary     string       `gorm:"type:text"`
	TotalTokens    uint         `gorm:"type:integer;not null"`
	Visible        bool         `gorm:"type:bool;not null"`
	Type           MessageType  `gorm:"type:smallint;not null; default:1"`
}
