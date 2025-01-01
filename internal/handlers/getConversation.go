package handlers

import (
	"net/http"
	"smart-chat/internal/models"
	"smart-chat/internal/services/conversation"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetConversationHandler(conversationService *conversation.ConversationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		session, exists := c.Get("session")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized - No session found"})
			return
		}

		authSession, ok := session.(models.Session)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error - Session casting issue"})
			return
		}

		authSessionWithConversations, err := conversationService.GetSessionWithConversations(authSession.ID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			}
			return
		}

		if len(authSessionWithConversations.Conversations) == 0 {
			c.JSON(http.StatusOK, gin.H{"error": "No conversations found for this session"})
			return
		}

		// Assuming we're interested in the first conversation's history
		firstConversation := authSessionWithConversations.Conversations[0]
		formattedHistory := []gin.H{}
		for _, messagePair := range firstConversation.MessagePairs {
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
