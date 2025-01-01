package models

import (
	"gorm.io/gorm"
)

type ConvAnalysis struct {
	gorm.Model
	ConversationID uint         `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Conversation   Conversation `gorm:"foreignKey:ConversationID;references:ID"`
	Summary        string       `gorm:"type:text"`
	EmailSent      bool         `gorm:"type:bool;not null;default:0"`
}
