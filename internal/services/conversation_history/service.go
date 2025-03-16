package convHistory

import (
	"smart-chat/internal/models"
	spec "smart-chat/internal/services/conversation_history/specification"

	"gorm.io/gorm"
)

type ConvHistoryService struct {
	db *gorm.DB
}

func NewConvHistoryService(db *gorm.DB) *ConvHistoryService {
	return &ConvHistoryService{db: db}
}

// GetConversations fetches conversations with optional specs, offset, and limit.
func (chs *ConvHistoryService) GetConversations(offset, limit int, specs ...spec.Specification) ([]models.Conversation, error) {
	dbQuery := chs.db.Model(&models.Conversation{})

	// Apply each specification to the query
	for _, s := range specs {
		dbQuery = s.Apply(dbQuery)
	}

	var conversations []models.Conversation
	err := dbQuery.
		Preload("MessagePairs").
		Preload("FunctionCalls").
		Preload("Session.User").
		Offset(offset).
		Limit(limit).
		Find(&conversations).Error

	if err != nil {
		return nil, err
	}

	return conversations, nil
}

// CountConversations returns the total number of conversations matching specs.
func (chs *ConvHistoryService) CountConversations(specs ...spec.Specification) (int64, error) {
	dbQuery := chs.db.Model(&models.Conversation{})

	// Apply each specification to the query
	for _, s := range specs {
		dbQuery = s.Apply(dbQuery)
	}

	var count int64
	if err := dbQuery.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
