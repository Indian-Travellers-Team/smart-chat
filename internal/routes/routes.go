package routes

import (
	"smart-chat/internal/authservice/zitadel"
	"smart-chat/external/indian_travellers"
	"smart-chat/internal/handlers"
	"smart-chat/internal/services/analytics"
	authUserConversation "smart-chat/internal/services/auth_user_conversation"
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

	group.POST("/start", handlers.StartConversationHandler(convService, jobService, slackService))
	group.GET("/messages", handlers.GetConversationHandler(convService))

	group.POST("/message",
		handlers.RespondConversationHandler(convService, jobService, slackService))
}

func ClientRoutes(
	group *gin.RouterGroup,
	convHistoryService *convHistory.ConvHistoryService,
	analyticsService *analytics.AnalyticsService,
	us *userService.UserService,
	humanService *human.HumanService,
	jobService *notifications_job.JobService,
	slackService *slack.SlackService,
	authUserConversationService *authUserConversation.Service,
	tokenValidator zitadel.TokenValidator,
) {
	group.POST("/login", handlers.ClientAdminLoginHandler())
	group.GET("/conversation/:id", handlers.GetConversationByIDHandler(convHistoryService))
	group.GET("/conversations", handlers.GetConversationsWithFiltersHandler(convHistoryService))
	group.GET("/analytics/dashboard/conversations-summary", handlers.GetDashboardConversationSummaryHandler(analyticsService))
	group.GET("/analytics/conversations/last-30-days", handlers.GetConversationsCountLast30DaysHandler(analyticsService))
	group.GET("/userdetails", handlers.ClientUserDetailsHandler(us))
	group.POST("/add-message", handlers.AddMessageHandler(humanService, jobService, slackService))
	group.POST("/conversations/link", handlers.LinkAuthUserConversationsHandler(authUserConversationService, tokenValidator))
}
