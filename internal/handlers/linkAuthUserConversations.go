package handlers

import (
	"net/http"
	"strings"

	"smart-chat/internal/authservice/zitadel"
	authUserConversation "smart-chat/internal/services/auth_user_conversation"

	"github.com/gin-gonic/gin"
)

type LinkAuthUserConversationsRequest struct {
	UserID          uint   `json:"user_id" binding:"required"`
	ConversationIDs []uint `json:"conversation_ids" binding:"required"`
}

func LinkAuthUserConversationsHandler(
	service *authUserConversation.Service,
	tokenValidator zitadel.TokenValidator,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LinkAuthUserConversationsRequest
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

		isAdmin, err := service.IsAdminByZitadelUserID(*validatedUser.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify admin role"})
			return
		}

		if !isAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin role required"})
			return
		}

		linkedCount, err := service.LinkConversations(req.UserID, req.ConversationIDs)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":       "linked",
			"linked_count": linkedCount,
		})
	}
}

func tokenFromAuthorizationHeader(authHeader string) string {
	authHeader = strings.TrimSpace(authHeader)
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}

	// Backward compatible with existing APIs that pass raw token directly.
	return authHeader
}
