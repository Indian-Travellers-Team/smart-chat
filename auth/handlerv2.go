// auth/handler.go
package auth

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func InitLoginHandlerv2(authService *AuthV2Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var info UserLoginInfo
		if err := c.ShouldBindJSON(&info); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Both name and mobile are required"})
			return
		}

		token, err := authService.InitLogin(info)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate login"})
			return
		}

		c.JSON(http.StatusOK, token)
	}
}

func WARefreshTokenHandler(authService *AuthV2Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var info RefreshTokenInfo
		if err := c.ShouldBindJSON(&info); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
			return
		}
		accessToken := c.GetHeader("Authorization")
		if accessToken == "" {
			log.Println("Authorization header missing")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization Required"})
			c.Abort()
			return
		}
		token, err := authService.RefreshToken(info, accessToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login"})
			return
		}

		c.JSON(http.StatusOK, token)
	}
}

func WALoginHandler(authService *AuthV2Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var info UserLoginInfoWA
		if err := c.ShouldBindJSON(&info); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
			return
		}

		token, err := authService.NewLoginWA(info)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login"})
			return
		}

		c.JSON(http.StatusOK, token)
	}
}

func ValidateLoginHandlerv2(authService *AuthV2Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginValidationInfo
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Both OTP and Token are required", "success": false})
			return
		}

		accessToken, err := authService.ValidateLogin(req.Token, req.OTP)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate login", "success": false})
			return
		}

		c.JSON(http.StatusOK, accessToken)
	}
}
