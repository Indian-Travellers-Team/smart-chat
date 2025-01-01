// auth/routes.go
package auth

import (
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(group *gin.RouterGroup, authService *AuthService) {
	group.POST("/init-login", InitLoginHandler(authService))
	group.POST("/validate-login", ValidateLoginHandler(authService))
}

func RegisterV2AuthRoutes(group *gin.RouterGroup, authService *AuthV2Service) {
	group.POST("/init-login", InitLoginHandlerv2(authService))
	group.POST("/validate-login", ValidateLoginHandlerv2(authService))
	group.POST("/login-for-whatsapp", WALoginHandler(authService))
}
