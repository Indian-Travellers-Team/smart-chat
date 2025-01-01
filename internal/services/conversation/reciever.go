package conversation

import (
	"smart-chat/internal/models"

	"gorm.io/gorm"
)

type ConversationReceiver struct {
	db            *gorm.DB
	Builder       *ConversationBuilder
	Executor      *ConversationExecutor
	ConvState     *ConversationState
	HistoryLoader *ConversationHistory
}

func NewConversationReceiver(db *gorm.DB, builder *ConversationBuilder, executor *ConversationExecutor, state *ConversationState, historyLoader *ConversationHistory) *ConversationReceiver {
	return &ConversationReceiver{db: db, Builder: builder, Executor: executor, ConvState: state, HistoryLoader: historyLoader}
}

func (cr *ConversationReceiver) ReceiveMessage(sessionID uint, message string, messageType models.MessageType) (string, error) {
	conversation, err := cr.Builder.Build(sessionID)
	if err != nil {
		return "", err
	}
	convHistory, _ := cr.HistoryLoader.FetchHistory(conversation.ID)
	cr.ConvState.InitState(conversation.ID, convHistory)
	response, err := cr.Executor.Execute(conversation.ID, message, messageType, cr.ConvState)
	if err != nil {
		return "", err
	}
	return response, nil
}
