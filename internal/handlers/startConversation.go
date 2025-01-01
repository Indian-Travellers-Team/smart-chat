package handlers

import (
	"net/http"
	"smart-chat/internal/models"
	"smart-chat/internal/services/conversation"

	"github.com/gin-gonic/gin"
)

func StartConversationHandler(conversationService *conversation.ConversationService) gin.HandlerFunc {
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

		// Here, we're using a hardcoded "Hello" message. In a real application, you'd likely get this from the request.
		userInput := "Hello!"

		// Handle the session/message using the ConversationService. Here, authUser.ID could be used to find or start a session.
		response, err := conversationService.HandleSession(authSession.ID, userInput, models.MessageTypeUserFix)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to handle session"})
			return
		}

		// Response to indicate the conversation has been handled/started.
		// In a real scenario, you might want to send back a more meaningful response.
		c.JSON(http.StatusOK, gin.H{"response": response})
	}
}
