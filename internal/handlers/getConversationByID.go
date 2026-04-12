package handlers

import (
	"log"
	"net/http"
	"strings"

	"smart-chat/internal/authservice/zitadel"
	convHistory "smart-chat/internal/services/conversation_history"
	"smart-chat/internal/services/conversation_history/specification"
	authUserConversation "smart-chat/internal/services/auth_user_conversation"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetConversationByIDHandler(
	historyService *convHistory.ConvHistoryService,
	authUserConversationService *authUserConversation.Service,
	tokenValidator zitadel.TokenValidator,
) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		principal, err := authUserConversationService.GetAuthPrincipalByZitadelUserID(*validatedUser.ID)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "auth user not found"})
			return
		}

		roleName := strings.ToUpper(strings.TrimSpace(principal.RoleName))
		if roleName != "ADMIN" && roleName != "AGENT" {
			c.JSON(http.StatusForbidden, gin.H{"error": "agent or admin role required"})
			return
		}

		// Extract ID from path parameters.
		idParam := c.Param("id")
		id, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			log.Printf("Invalid ID format: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID format"})
			return
		}

		specs := []specification.Specification{specification.ByID{ID: uint(id)}}
		if roleName != "ADMIN" {
			specs = append(specs, specification.ByAssignedAuthUser{AuthUserID: principal.UserID})
		}

		// Since we only want one conversation, pass offset=0 and limit=1.
		conversations, err := historyService.GetConversations(0, 1, specs...)
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
