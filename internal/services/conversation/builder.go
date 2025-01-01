package conversation

import (
	"smart-chat/internal/models"

	"gorm.io/gorm"
)

type ConversationBuilder struct {
	db *gorm.DB
}

func NewConversationBuilder(db *gorm.DB) *ConversationBuilder {
	return &ConversationBuilder{db: db}
}

func (cb *ConversationBuilder) Build(sessionID uint) (*models.Conversation, error) {
	conversation := &models.Conversation{
		SessionID: sessionID,
	}
	if result := cb.db.FirstOrCreate(conversation, models.Conversation{SessionID: sessionID}); result.Error != nil {
		return nil, result.Error
	}
	return conversation, nil
}
