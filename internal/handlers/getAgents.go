package handlers

import (
	"net/http"
	"strings"

	"smart-chat/internal/authservice/zitadel"
	authUserConversation "smart-chat/internal/services/auth_user_conversation"

	"github.com/gin-gonic/gin"
)

func GetAgentsHandler(
	service *authUserConversation.Service,
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

		isAdmin, err := service.IsAdminByZitadelUserID(*validatedUser.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify admin role"})
			return
		}

		if !isAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin role required"})
			return
		}

		agents, err := service.ListAgentsAndAdmins()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch agents"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"agents": agents})
	}
}
