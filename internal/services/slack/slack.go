package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"smart-chat/internal/models"

	"gorm.io/gorm"
)

// SlackService encapsulates Slack endpoint URLs for notifications and alerts, along with a DB instance.
type SlackService struct {
	NotificationURL string
	AlertURL        string
	db              *gorm.DB
}

// NewSlackService returns a new SlackService.
func NewSlackService(notificationURL, alertURL string, db *gorm.DB) *SlackService {
	return &SlackService{
		NotificationURL: notificationURL,
		AlertURL:        alertURL,
		db:              db,
	}
}

// SendSlackNotification sends a notification message to Slack synchronously.
// It returns a boolean indicating success and an error if any.
func (s *SlackService) SendSlackNotification(message string) (bool, error) {
	payload := map[string]string{
		"text": fmt.Sprintf("service:indian_travellers_cms -- %s", message),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return false, err
	}

	resp, err := http.Post(s.NotificationURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return true, nil
	}
	return false, nil
}

// SendSlackAlert sends an alert message to Slack synchronously.
// It returns a boolean indicating success and an error if any.
func (s *SlackService) SendSlackAlert(message string) (bool, error) {
	payload := map[string]string{
		"text": fmt.Sprintf("service:indian_travellers_cms -- %s", message),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return false, err
	}

	resp, err := http.Post(s.AlertURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return true, nil
	}
	return false, nil
}

// SendSlackNotificationAsync sends a notification message to Slack in a background job.
// It spawns a goroutine that calls the synchronous version. Any errors are logged.
func (s *SlackService) SendSlackNotificationAsync(message string) {
	go func() {
		success, err := s.SendSlackNotification(message)
		if err != nil || !success {
			log.Printf("failed to send Slack notification: %v", err)
		}
	}()
}

// SendSlackAlertAsync sends an alert message to Slack in a background job.
// It spawns a goroutine that calls the synchronous version. Any errors are logged.
func (s *SlackService) SendSlackAlertAsync(message string) {
	go func() {
		success, err := s.SendSlackAlert(message)
		if err != nil || !success {
			log.Printf("failed to send Slack alert: %v", err)
		}
	}()
}

// NotifyNewConversation fetches the most recent conversation for the given session
// and sends a Slack notification asynchronously. It also includes whether the conversation is via WhatsApp.
func (s *SlackService) NotifyNewConversation(session models.Session, whatsapp bool) {
	go func() {
		var conv models.Conversation
		err := s.db.
			Preload("Session.User").
			Where("session_id = ?", session.ID).
			Order("created_at desc").
			First(&conv).Error
		if err != nil {
			log.Printf("failed to find conversation for session %d: %v", session.ID, err)
			return
		}
		message := ""
		if whatsapp {
			message = fmt.Sprintf("New conversation with ID: *%d* started in WhatsApp ", conv.ID)
		} else {
			message = fmt.Sprintf("New conversation with ID: *%d* started in Website ", conv.ID)
		}
		s.SendSlackNotificationAsync(message)
	}()
}
