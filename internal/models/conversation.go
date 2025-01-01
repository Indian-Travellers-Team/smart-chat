package models

import (
	"gorm.io/gorm"
)

type Conversation struct {
	gorm.Model
	SessionID     uint    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Session       Session `gorm:"foreignKey:SessionID;references:ID"`
	Feedback      bool    `gorm:"type:bool;null"`
	TotalTokens   int
	MessagePairs  []MessagePair
	FunctionCalls []FunctionCall
	Analysed      bool `gorm:"type:bool;not null;default=0"`
}
