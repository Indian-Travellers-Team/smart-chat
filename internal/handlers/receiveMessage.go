package handlers

import (
	"smart-chat/internal/store"

	"github.com/gin-gonic/gin"
)

func ReceiveMessageHandler(c *gin.Context) {
	accessToken := c.GetHeader("Authorization")
	conversation, _ := store.GetConversation(accessToken)

	// Format the conversation history for the response
	formattedHistory := make([]gin.H, 0, len(conversation.Messages))
	for _, messagePair := range conversation.Messages {
		formattedHistory = append(formattedHistory, gin.H{
			"UserMessage": messagePair.UserMessage,
			"BotMessage":  messagePair.BotMessage,
		})
	}

	c.JSON(200, gin.H{"conversationHistory": formattedHistory})
}
