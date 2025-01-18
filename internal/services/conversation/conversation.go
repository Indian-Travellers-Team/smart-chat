package conversation

import (
	"smart-chat/internal/models"

	"gorm.io/gorm"
)

type ConversationService struct {
	DB       *gorm.DB
	Receiver *ConversationReceiver
}

func NewConversationService(db *gorm.DB) *ConversationService {
	builder := NewConversationBuilder(db)
	executor := NewConversationExecutor(db)
	state := NewConversationState(db)
	historyLoader := NewConversationHistory(db)
	receiver := NewConversationReceiver(db, builder, executor, state, historyLoader)
	return &ConversationService{
		DB:       db,
		Receiver: receiver,
	}
}

func (cs *ConversationService) HandleSession(sessionID uint, userInput string, messageType models.MessageType, whatsapp bool) (string, error) {
	return cs.Receiver.ReceiveMessage(sessionID, userInput, messageType, whatsapp)
}

func (cs *ConversationService) GetSessionWithConversations(sessionID uint) (*models.Session, error) {
	var session models.Session
	err := cs.DB.Preload("Conversations.MessagePairs").Where("id = ?", sessionID).First(&session).Error
	return &session, err
}
