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

func (chs *ConvHistoryService) GetConversations(specs ...spec.Specification) ([]models.Conversation, error) {

	dbQuery := chs.db.Model(&models.Conversation{})

	// Apply each specification to the query
	for _, spec := range specs {
		dbQuery = spec.Apply(dbQuery)
	}

	var conversations []models.Conversation
	err := dbQuery.Preload("MessagePairs").Preload("FunctionCalls").Preload("Session.User").Find(&conversations).Error
	if err != nil {
		return nil, err
	}

	return conversations, nil
}
