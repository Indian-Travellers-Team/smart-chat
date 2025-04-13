package human

import (
	"fmt"

	"smart-chat/internal/models"

	"gorm.io/gorm"
)

// HumanService is responsible for adding agent messages (as if by a human)
// to a conversation. It works directly with the database.
type HumanService struct {
	db *gorm.DB
}

// NewHumanService returns a new instance of HumanService.
func NewHumanService(db *gorm.DB) *HumanService {
	return &HumanService{db: db}
}

// AddMessage finds the conversation by ID and adds a new MessagePair
// with Type set to MessageTypeAgentAssumedAssistant and User left empty.
// It stores the message in JSON format: {"content": "your message"}
func (hs *HumanService) AddMessage(conversationID uint, message string) error {
	// 1. Ensure the conversation exists.
	var conv models.Conversation
	if err := hs.db.First(&conv, conversationID).Error; err != nil {
		return fmt.Errorf("conversation not found (ID=%d): %w", conversationID, err)
	}

	// 2. Wrap the incoming message in a JSON object.
	//    This produces a string like: {"content":"Great choice, Deepak! ..."}
	formattedMessage := fmt.Sprintf("{\"content\":%q}", message)

	// 3. Build a new MessagePair record.
	msgPair := models.MessagePair{
		ConversationID: conversationID,
		User:           "", // Keep user field empty.
		Bot:            formattedMessage,
		Visible:        true,
		Type:           models.MessageTypeAgentAssumedAssistant,
		TotalTokens:    0, // Adjust if necessary.
	}

	// 4. Insert the message pair record.
	if err := hs.db.Create(&msgPair).Error; err != nil {
		return fmt.Errorf("failed to add message: %w", err)
	}

	return nil
}
