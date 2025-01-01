package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ClientLoginInfo struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func ClientAdminLoginHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var info ClientLoginInfo
		if err := c.ShouldBindJSON(&info); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Both username and password are required"})
			return
		}
		log.Printf("username %v", info.Username)
		log.Printf("password %v", info.Password)
		token, err := generateSecretToken(32)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate login"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"authToken": token})
	}
}

func generateSecretToken(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
