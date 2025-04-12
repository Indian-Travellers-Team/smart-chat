package handlers

import (
	"net/http"

	"smart-chat/internal/models"
	"smart-chat/internal/services/conversation"
	"smart-chat/internal/services/notifications_job"
	"smart-chat/internal/services/slack"

	"github.com/gin-gonic/gin"
)

func RespondConversationHandler(
	convService *conversation.ConversationService,
	jobService *notifications_job.JobService,
	slackService *slack.SlackService,
) gin.HandlerFunc {
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

		var reqBody struct {
			Message string `json:"message" binding:"required"`
		}
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		whatsapp := c.DefaultQuery("whatsapp", "false") == "true"
		userInput := reqBody.Message

		// 1. Handle the conversation.
		response, err := convService.HandleSession(
			authSession.ID,
			userInput,
			models.MessageTypeUserSent,
			whatsapp,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to handle session"})
			return
		}
		if whatsapp {
			// 2. Run the notification job in the background.
			go jobService.SendConversationNotification(userInput, response, authSession, slackService)
		}

		// 3. Return the conversation response.
		c.JSON(http.StatusOK, gin.H{"response": response})
	}
}
