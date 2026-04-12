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

// GetConversationsWithSort fetches conversations with full associations (MessagePairs, FunctionCalls, Session.User).
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

// ListConversations fetches conversations for list views.
// Only the columns required for the list response are selected; MessagePairs and FunctionCalls
// are intentionally omitted to avoid expensive N+1 loads.
func (chs *ConvHistoryService) ListConversations(offset, limit int, sortOrder string, specs ...spec.Specification) ([]models.Conversation, error) {
	sortOrder = strings.ToLower(sortOrder)
	if sortOrder != constants.SortAsc && sortOrder != constants.SortDesc {
		sortOrder = constants.DefaultSortStr
	}

	dbQuery := chs.db.Model(&models.Conversation{})

	for _, s := range specs {
		dbQuery = s.Apply(dbQuery)
	}

	var conversations []models.Conversation
	err := dbQuery.
		Distinct("conversations.id", "conversations.created_at", "conversations.session_id").
		Select("conversations.id", "conversations.created_at", "conversations.session_id").
		Preload("Session", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "source", "user_id")
		}).
		Preload("Session.User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name", "mobile")
		}).
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
	if err := dbQuery.Distinct("conversations.id").Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
