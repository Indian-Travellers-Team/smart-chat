package handlers

import (
	"net/http"

	"smart-chat/internal/services/human"
	"smart-chat/internal/services/notifications_job"
	"smart-chat/internal/services/slack"

	"github.com/gin-gonic/gin"
)

// AddMessageHandler handles POST /add-message requests.
// It expects a JSON body containing conversation_id and message.
func AddMessageHandler(hs *human.HumanService, jobService *notifications_job.JobService, slackServie *slack.SlackService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ConversationID uint   `json:"conversation_id" binding:"required"`
			Message        string `json:"message" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := hs.AddMessage(req.ConversationID, req.Message); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		go jobService.SendConversationNotificationByID("", req.Message, req.ConversationID, slackServie)

		c.JSON(http.StatusOK, gin.H{"status": "Message added successfully"})
	}
}
