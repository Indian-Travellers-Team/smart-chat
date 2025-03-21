package routes

import (
	"smart-chat/external/notification"
	"smart-chat/internal/handlers"
	"smart-chat/internal/services/conversation"
	convHistory "smart-chat/internal/services/conversation_history"
	"smart-chat/internal/services/notifications_job"
	userService "smart-chat/internal/services/user"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/ping", handlers.PingHandler)
	group.POST("/send", handlers.SendMessageHandler)
	group.GET("/receive", handlers.ReceiveMessageHandler)
}

func RegisterV2Routes(group *gin.RouterGroup, convService *conversation.ConversationService) {

	notifClient := notification.NewClient("http://notification_service")

	jobService := notifications_job.NewJobService(notifClient)

	group.POST("/start", handlers.StartConversationHandler(convService))
	group.GET("/messages", handlers.GetConversationHandler(convService))

	group.GET("/start",
		handlers.RespondConversationHandler(convService, jobService))
}

func ClientRoutes(group *gin.RouterGroup, convHistoryService *convHistory.ConvHistoryService, us *userService.UserService) {
	group.POST("/login", handlers.ClientAdminLoginHandler())
	group.GET("/conversation/:id", handlers.GetConversationByIDHandler(convHistoryService))
	group.GET("/conversations", handlers.GetConversationsWithFiltersHandler(convHistoryService))
	group.GET("/userdetails", handlers.ClientUserDetailsHandler(us))
}
