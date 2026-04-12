package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"smart-chat/internal/authservice/zitadel"
	authUserConversation "smart-chat/internal/services/auth_user_conversation"
	"smart-chat/internal/services/slack"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UpdateAuthUserConversationRequest struct {
	ConversationID uint    `json:"conversation_id" binding:"required"`
	Started        *bool   `json:"started"`
	Resolved       *bool   `json:"resolved"`
	Comments       *string `json:"comments"`
}

func UpdateAuthUserConversationHandler(
	service *authUserConversation.Service,
	tokenValidator zitadel.TokenValidator,
	slackService *slack.SlackService,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UpdateAuthUserConversationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		rawToken := tokenFromAuthorizationHeader(c.GetHeader("Authorization"))
		if rawToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization Required"})
			return
		}

		validatedUser, err := tokenValidator.ValidateToken(c.Request.Context(), rawToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid access token"})
			return
		}

		if validatedUser == nil || validatedUser.ID == nil || strings.TrimSpace(*validatedUser.ID) == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid access token"})
			return
		}

		principal, err := service.GetAuthPrincipalByZitadelUserID(*validatedUser.ID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusForbidden, gin.H{"error": "auth user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve auth user"})
			return
		}

		roleName := strings.ToUpper(strings.TrimSpace(principal.RoleName))
		if roleName != "ADMIN" && roleName != "AGENT" {
			c.JSON(http.StatusForbidden, gin.H{"error": "agent or admin role required"})
			return
		}

		updatedLink, err := service.UpdateConversationTracking(authUserConversation.UpdateConversationTrackingInput{
			AuthUserID:     principal.UserID,
			ConversationID: req.ConversationID,
			Started:        req.Started,
			Resolved:       req.Resolved,
			Comments:       req.Comments,
		})
		if err != nil {
			switch {
			case errors.Is(err, gorm.ErrRecordNotFound):
				c.JSON(http.StatusNotFound, gin.H{"error": "conversation assignment not found for agent"})
			case err.Error() == "at least one field must be provided for update":
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update conversation tracking"})
			}
			return
		}

		if slackService != nil {
			slackService.SendSlackNotificationAsync(buildConversationTrackingSlackMessage(updatedLink))
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "updated",
			"agent": gin.H{
				"user_id": updatedLink.AuthUserID,
				"name":    updatedLink.AgentName,
			},
			"conversation": gin.H{
				"id":       updatedLink.ConversationID,
				"started":  updatedLink.Started,
				"resolved": updatedLink.Resolved,
				"comments": updatedLink.Comments,
			},
		})
	}
}

func buildConversationTrackingSlackMessage(link *authUserConversation.ConversationTracking) string {
	agentName := fmt.Sprintf("user_id=%d", link.AuthUserID)
	if link.AgentName != nil && strings.TrimSpace(*link.AgentName) != "" {
		agentName = fmt.Sprintf("%s (user_id=%d)", strings.TrimSpace(*link.AgentName), link.AuthUserID)
	}

	return fmt.Sprintf(
		"Conversation tracking updated for conversation ID: *%d* by *%s* | started=%t | resolved=%t | comments=%q",
		link.ConversationID,
		agentName,
		link.Started,
		link.Resolved,
		link.Comments,
	)
}