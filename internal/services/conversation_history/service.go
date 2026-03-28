package convHistory

import (
	"strings"

	"smart-chat/internal/constants"
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
	return chs.GetConversationsWithSort(offset, limit, constants.DefaultSortStr, specs...)
}

// GetConversationsWithSort fetches conversations with optional specs, offset, limit, and sort order.
func (chs *ConvHistoryService) GetConversationsWithSort(offset, limit int, sortOrder string, specs ...spec.Specification) ([]models.Conversation, error) {
	sortOrder = strings.ToLower(sortOrder)
	if sortOrder != constants.SortAsc && sortOrder != constants.SortDesc {
		sortOrder = constants.DefaultSortStr
	}

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
		Order("conversations.created_at " + sortOrder).
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
