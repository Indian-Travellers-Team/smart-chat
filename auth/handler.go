// auth/handler.go
package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserLoginInfo struct {
	Name   string `json:"name" binding:"required"`
	Mobile string `json:"mobile" binding:"required"`
}

type LoginValidationInfo struct {
	Token string `json:"token" binding:"required"`
	OTP   string `json:"otp" binding:"required"`
}

func InitLoginHandler(authService *AuthService) gin.HandlerFunc {
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

func ValidateLoginHandler(authService *AuthService) gin.HandlerFunc {
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
