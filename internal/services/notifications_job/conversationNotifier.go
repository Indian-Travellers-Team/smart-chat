package notifications_job

import (
	"encoding/json"
	"fmt"
	"log"

	"smart-chat/external/notification"
	"smart-chat/internal/models"
	"smart-chat/internal/services/slack"

	"gorm.io/gorm"
)

type JobService struct {
	notifClient *notification.Client
	db          *gorm.DB
}

func NewJobService(client *notification.Client, db *gorm.DB) *JobService {
	return &JobService{
		notifClient: client,
		db:          db,
	}
}

// SendConversationNotification is the background job that sends a conversation notification.
func (js *JobService) SendConversationNotification(userInput, botResponse string, session models.Session, slackService *slack.SlackService) {
	var parsed struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(botResponse), &parsed); err != nil {
		log.Printf("failed to parse bot response: %v", err)
		parsed.Content = botResponse
	}

	// Query the most recent conversation associated with this session.
	var conv models.Conversation
	err := js.db.
		Where("session_id = ?", session.ID).
		Order("created_at desc").
		First(&conv).Error
	if err != nil {
		log.Printf("failed to find conversation for session %d: %v", session.ID, err)
		// Optionally, handle error or set conv.ID to a fallback value.
		return
	}

	payload := notification.Payload{
		ConversationID: conv.ID,
		Mobile:         session.User.Mobile,
		MessagePair: notification.MessagePair{
			User: userInput,
			Bot:  parsed.Content,
		},
	}

	if err := js.notifClient.SendMessageEvent(payload); err != nil {
		slackService.SendSlackAlertAsync(fmt.Sprintf("failed to send notification: *%v* for conversation id *%d*", err, conv.ID))
		log.Printf("failed to send notification: %v", err)
	}
}

// SendConversationNotificationByID is an alternative background job that sends a conversation notification
// based on a given conversationID rather than a session.
func (js *JobService) SendConversationNotificationByID(userInput, botResponse string, conversationID uint, slackService *slack.SlackService) {
	// Query the conversation based on the provided conversationID, preloading Session and its User.
	var conv models.Conversation
	err := js.db.
		Preload("Session.User").
		First(&conv, conversationID).Error
	if err != nil {
		log.Printf("failed to find conversation with ID %d: %v", conversationID, err)
		return
	}

	payload := notification.Payload{
		ConversationID: conv.ID,
		Mobile:         conv.Session.User.Mobile,
		MessagePair: notification.MessagePair{
			User: userInput,
			Bot:  botResponse,
		},
	}

	if err := js.notifClient.SendMessageEvent(payload); err != nil {
		slackService.SendSlackAlertAsync(fmt.Sprintf("failed to send notification: *%v* for conversation id *%d*", err, conversationID))
		log.Printf("failed to send notification: %v", err)
	}
}
