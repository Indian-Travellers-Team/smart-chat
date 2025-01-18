package handlers

import (
	"net/http"
	"smart-chat/internal/models"
	"smart-chat/internal/services/conversation"

	"github.com/gin-gonic/gin"
)

func RespondConversationHandler(conversationService *conversation.ConversationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		session, exists := c.Get("session")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized - No user found"})
			return
		}

		authSession, ok := session.(models.Session)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error - User casting issue"})
			return
		}

		var jsonData struct {
			Message string `json:"message" binding:"required"`
		}

		if err := c.ShouldBindJSON(&jsonData); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})

		}

		// Get the whatsapp parameter, default to false
		whatsapp := c.DefaultQuery("whatsapp", "false") == "true"

		userInput := jsonData.Message

		response, err := conversationService.HandleSession(authSession.ID, userInput, models.MessageTypeUserSent, whatsapp)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to handle session"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"response": response})
	}
}
