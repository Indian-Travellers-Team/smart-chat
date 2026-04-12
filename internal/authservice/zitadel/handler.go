package zitadel

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func ValidateTokenHandler(validator TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ValidateTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil && !isEOFError(err) {
			errMsg := "invalid request payload"
			c.JSON(http.StatusOK, ValidateTokenResponse{User: nil, Error: &errMsg})
			return
		}

		rawToken := resolveToken(req.Token, c.GetHeader("Authorization"))
		if rawToken == "" {
			errMsg := "token is required"
			c.JSON(http.StatusOK, ValidateTokenResponse{User: nil, Error: &errMsg})
			return
		}

		user, err := validator.ValidateToken(c.Request.Context(), rawToken)
		if err != nil {
			errMsg := err.Error()
			c.JSON(http.StatusOK, ValidateTokenResponse{User: nil, Error: &errMsg})
			return
		}

		c.JSON(http.StatusOK, ValidateTokenResponse{User: user, Error: nil})
	}
}

func RegisterRoutes(group *gin.RouterGroup, validator TokenValidator) {
	group.POST("/token/validate", ValidateTokenHandler(validator))
}

func resolveToken(bodyToken *string, authHeader string) string {
	if bodyToken != nil {
		if token := strings.TrimSpace(*bodyToken); token != "" {
			return token
		}
	}

	authHeader = strings.TrimSpace(authHeader)
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}

	return ""
}

func isEOFError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "eof")
}
