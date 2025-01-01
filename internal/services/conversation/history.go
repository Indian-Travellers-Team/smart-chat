package conversation

import (
	"log"
	"smart-chat/internal/models"

	openai "github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
)

type ConversationHistory struct {
	db *gorm.DB
}

func NewConversationHistory(db *gorm.DB) *ConversationHistory {
	return &ConversationHistory{db: db}
}

func (ch *ConversationHistory) FetchHistory(conversationID uint) ([]openai.ChatCompletionMessage, error) {
	var conversationHistory []models.MessagePair
	err := ch.db.Where("conversation_id = ?", conversationID).Find(&conversationHistory).Error
	if err != nil {
		log.Printf("Error fetching conversation history: %v", err)
		return nil, err
	}

	messages := make([]openai.ChatCompletionMessage, 0, len(conversationHistory)+1)
	for _, pair := range conversationHistory {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: pair.User,
		}, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: pair.Bot,
		})
	}

	return messages, nil
}
