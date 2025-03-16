package handlers

import (
	"log"
	"net/http"
	convHistory "smart-chat/internal/services/conversation_history"
	"smart-chat/internal/services/conversation_history/specification"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetConversationByIDHandler(historyService *convHistory.ConvHistoryService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract ID from path parameters.
		idParam := c.Param("id")
		id, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			log.Printf("Invalid ID format: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID format"})
			return
		}

		// Create specification for fetching by ID.
		idSpec := specification.ByID{ID: uint(id)}

		// Since we only want one conversation, pass offset=0 and limit=1.
		conversations, err := historyService.GetConversations(0, 1, idSpec)
		if err != nil {
			log.Printf("Error fetching conversations: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching conversation history"})
			return
		}

		if len(conversations) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
			return
		}

		conversation := conversations[0]
		formattedHistory := make([]gin.H, 0)
		for _, messagePair := range conversation.MessagePairs {
			if messagePair.Visible {
				formattedHistory = append(formattedHistory, gin.H{
					"UserMessage": messagePair.User,
					"BotMessage":  messagePair.Bot,
				})
			}
		}

		c.JSON(http.StatusOK, gin.H{"conversationHistory": formattedHistory})
	}
}
