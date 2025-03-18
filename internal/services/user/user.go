package user

import (
	"time"

	"smart-chat/internal/models"

	"gorm.io/gorm"
)

// UserDetails represents the user details information to return.
type UserDetails struct {
	Name            string `json:"name"`
	Mobile          string `json:"mobile"`
	Source          string `json:"source"`
	LastMessageTime string `json:"lastMessageTime"`
	TotalMessages   int    `json:"totalMessages"`
}

// UserService provides methods for retrieving user-related data.
type UserService struct {
	db *gorm.DB
}

// NewUserService creates a new instance of UserService.
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// GetUserDetailsByConversationID retrieves user details based on a conversation ID.
// It preloads the associated Session.User and MessagePairs, then computes and returns:
// - the user's name,
// - mobile,
// - session source,
// - the last message time (formatted as a human readable string),
// - and the total number of messages in the conversation.
func (us *UserService) GetUserDetailsByConversationID(convID uint) (UserDetails, error) {
	var conv models.Conversation
	err := us.db.
		Preload("Session.User").
		Preload("MessagePairs").
		First(&conv, convID).Error
	if err != nil {
		return UserDetails{}, err
	}

	details := UserDetails{
		Name:   conv.Session.User.Name,
		Mobile: conv.Session.User.Mobile,
		Source: conv.Session.Source,
	}

	totalMessages := len(conv.MessagePairs)
	details.TotalMessages = totalMessages

	// Find the latest UpdatedAt among all message pairs.
	var lastMsgTime time.Time
	for _, mp := range conv.MessagePairs {
		if mp.UpdatedAt.After(lastMsgTime) {
			lastMsgTime = mp.UpdatedAt
		}
	}

	// Format the last message time in human readable format.
	// Using RFC3339Nano produces a string like "2025-02-23T03:55:17.222655+05:30".
	details.LastMessageTime = lastMsgTime.Format(time.RFC3339Nano)

	return details, nil
}
