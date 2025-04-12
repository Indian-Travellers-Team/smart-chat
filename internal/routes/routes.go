package routes

import (
	"smart-chat/external/indian_travellers"
	"smart-chat/internal/handlers"
	"smart-chat/internal/services/conversation"
	convHistory "smart-chat/internal/services/conversation_history"
	"smart-chat/internal/services/human"
	"smart-chat/internal/services/notifications_job"
	"smart-chat/internal/services/slack"
	userService "smart-chat/internal/services/user"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(group *gin.RouterGroup, indian_travellers *indian_travellers.Client) {
	group.GET("/ping", handlers.PingHandler)
	group.POST("/send", handlers.SendMessageHandler(indian_travellers))
	group.GET("/receive", handlers.ReceiveMessageHandler)
}

func RegisterV2Routes(
	group *gin.RouterGroup,
	convService *conversation.ConversationService,
	jobService *notifications_job.JobService,
	slackService *slack.SlackService,
) {

	group.POST("/start", handlers.StartConversationHandler(convService, slackService))
	group.GET("/messages", handlers.GetConversationHandler(convService))

	group.POST("/message",
		handlers.RespondConversationHandler(convService, jobService, slackService))
}

func ClientRoutes(
	group *gin.RouterGroup,
	convHistoryService *convHistory.ConvHistoryService,
	us *userService.UserService,
	humanService *human.HumanService,
	jobService *notifications_job.JobService,
	slackService *slack.SlackService,
) {
	group.POST("/login", handlers.ClientAdminLoginHandler())
	group.GET("/conversation/:id", handlers.GetConversationByIDHandler(convHistoryService))
	group.GET("/conversations", handlers.GetConversationsWithFiltersHandler(convHistoryService))
	group.GET("/userdetails", handlers.ClientUserDetailsHandler(us))
	group.POST("/add-message", handlers.AddMessageHandler(humanService, jobService, slackService))
}
