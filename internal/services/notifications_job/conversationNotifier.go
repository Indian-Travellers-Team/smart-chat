package notifications_job

import (
	"encoding/json"
	"log"

	"smart-chat/external/notification"
	"smart-chat/internal/models"

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
func (js *JobService) SendConversationNotification(userInput, botResponse string, session models.Session) {
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
		log.Printf("failed to send notification: %v", err)
	}
}
