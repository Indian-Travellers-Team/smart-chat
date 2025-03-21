package notifications_job

import (
	"encoding/json"
	"log"

	"smart-chat/external/notification"
	"smart-chat/internal/models"
)

type JobService struct {
	notifClient *notification.Client
}

func NewJobService(client *notification.Client) *JobService {
	return &JobService{notifClient: client}
}

// SendConversationNotification is the background job that sends a conversation notification.
func (js *JobService) SendConversationNotification(userInput, botResponse string, session models.Session) {
	var parsed struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(botResponse), &parsed); err != nil {
		log.Printf("failed to parse bot response: %v", err)
		parsed.Content = botResponse
	}

	// Example conversation ID logic:
	conversationID := uint(1)

	payload := notification.Payload{
		ConversationID: conversationID,
		Mobile:         session.User.Mobile,
		MessagePair: notification.MessagePair{
			User: userInput,
			Bot:  parsed.Content,
		},
	}

	if err := js.notifClient.SendMessageEvent(payload); err != nil {
		log.Printf("failed to send notification: %v", err)
	}
}
