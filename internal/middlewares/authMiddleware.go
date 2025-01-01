package middleware

import (
	"net/http"
	"smart-chat/internal/store"
	"time"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken := c.GetHeader("Authorization")
		if accessToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
			c.Abort()
			return
		}

		conversation, err := store.GetConversation(accessToken)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
			c.Abort()
			return
		}

		if time.Now().After(conversation.AccessTokenExpireTime) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "access token expired"})
			c.Abort()
			return
		}

		// Store the conversation in the context for use in the handler
		c.Set("conversation", conversation)
		c.Next()
	}
}
