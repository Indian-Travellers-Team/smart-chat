package conversation

import (
	"fmt"
	"log"
	"smart-chat/cache"
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

func (cr *ConversationReceiver) ReceiveMessage(sessionID uint, message string, messageType models.MessageType, whatsapp bool) (string, error) {
	conversation, err := cr.Builder.Build(sessionID)
	if err != nil {
		return "", err
	}
	convHistory, _ := cr.HistoryLoader.FetchHistory(conversation.ID)
	cr.ConvState.InitState(conversation.ID, convHistory)
	response, err := cr.Executor.Execute(conversation.ID, message, messageType, cr.ConvState, whatsapp)
	if err != nil {
		return "", err
	}
	// Build the cache key using the format defined in cache.CacheKeys.UserDetails.Key
	cacheKey := fmt.Sprintf(cache.CacheKeys.UserDetails.Key, conversation.ID)
	// Set the fetched details in cache with the TTL defined in cache.CacheKeys.UserDetails.TTL
	if err := cache.DeleteCache(cacheKey); err != nil {
		log.Println("Unable to delete cache for user details")
	}
	return response, nil
}
