package routes

import (
	"smart-chat/internal/handlers"
	"smart-chat/internal/services/conversation"
	convHistory "smart-chat/internal/services/conversation_history"
	userService "smart-chat/internal/services/user"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/ping", handlers.PingHandler)
	group.POST("/send", handlers.SendMessageHandler)
	group.GET("/receive", handlers.ReceiveMessageHandler)
}

func RegisterV2Routes(group *gin.RouterGroup, convService *conversation.ConversationService) {
	group.POST("/start", handlers.StartConversationHandler(convService))
	group.POST("/message", handlers.RespondConversationHandler(convService))
	group.GET("/messages", handlers.GetConversationHandler(convService))
}

func ClientRoutes(group *gin.RouterGroup, convHistoryService *convHistory.ConvHistoryService, us *userService.UserService) {
	group.POST("/login", handlers.ClientAdminLoginHandler())
	group.GET("/conversation/:id", handlers.GetConversationByIDHandler(convHistoryService))
	group.GET("/conversations", handlers.GetConversationsWithFiltersHandler(convHistoryService))
	group.GET("/userdetails", handlers.ClientUserDetailsHandler(us))
}
