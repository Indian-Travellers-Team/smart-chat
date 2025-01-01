package middleware

import (
	"net/http"
	"smart-chat/internal/models"
	"time"

	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthSessionMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken := c.GetHeader("Authorization")
		if accessToken == "" {
			log.Println("Authorization header missing")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization Required"})
			c.Abort()
			return
		}

		var session models.Session
		result := db.Preload("User").Where("auth_token = ?", accessToken).First(&session)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				log.Println("Session not found for token:", accessToken)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "session not found"})
			} else {
				log.Println("Database error:", result.Error)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}
			c.Abort()
			return
		}

		if time.Now().After(session.ExpireAt) {
			log.Println("Session expired for user:", session.User.ID)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "session expired"})
			c.Abort()
			return
		}

		c.Set("user", session.User)
		c.Set("session", session)
		c.Next()
	}
}
