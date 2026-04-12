package auth_user_conversation

import (
	"errors"
	"fmt"
	"strings"

	"smart-chat/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Service struct {
	db *gorm.DB
}

type AgentUser struct {
	UserID uint    `json:"user_id" gorm:"column:user_id"`
	Name   *string `json:"name" gorm:"column:name"`
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) LinkConversations(authUserID uint, conversationIDs []uint) (int, error) {
	if authUserID == 0 {
		return 0, errors.New("user_id is required")
	}

	if len(conversationIDs) == 0 {
		return 0, errors.New("conversation_ids cannot be empty")
	}

	var userCount int64
	if err := s.db.Table("auth_users").Where("user_id = ?", authUserID).Count(&userCount).Error; err != nil {
		return 0, err
	}
	if userCount == 0 {
		return 0, errors.New("auth user not found")
	}

	uniqueConversationIDs := dedupeUintIDs(conversationIDs)
	if len(uniqueConversationIDs) == 0 {
		return 0, errors.New("conversation_ids cannot be empty")
	}

	var foundIDs []uint
	if err := s.db.Model(&models.Conversation{}).Where("id IN ?", uniqueConversationIDs).Pluck("id", &foundIDs).Error; err != nil {
		return 0, err
	}

	if len(foundIDs) != len(uniqueConversationIDs) {
		return 0, fmt.Errorf("one or more conversation_ids are invalid")
	}

	links := make([]models.AuthUserConversation, 0, len(uniqueConversationIDs))
	for _, conversationID := range uniqueConversationIDs {
		links = append(links, models.AuthUserConversation{
			AuthUserID:     authUserID,
			ConversationID: conversationID,
		})
	}

	result := s.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "auth_user_id"},
			{Name: "conversation_id"},
		},
		DoNothing: true,
	}).Create(&links)
	if result.Error != nil {
		return 0, result.Error
	}

	return int(result.RowsAffected), nil
}

func (s *Service) IsAdminByZitadelUserID(zitadelUserID string) (bool, error) {
	zitadelUserID = strings.TrimSpace(zitadelUserID)
	if zitadelUserID == "" {
		return false, errors.New("zitadel user id is required")
	}

	type adminLookup struct {
		RoleName string `gorm:"column:name"`
	}

	var row adminLookup
	err := s.db.
		Table("auth_users").
		Select("auth_roles.name").
		Joins("JOIN auth_roles ON auth_roles.role_id = auth_users.role_id").
		Where("auth_users.zitadel_user_id = ?", zitadelUserID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	return strings.EqualFold(strings.TrimSpace(row.RoleName), "ADMIN"), nil
}

func (s *Service) ListAgentsAndAdmins() ([]AgentUser, error) {
	agents := make([]AgentUser, 0)

	err := s.db.
		Table("auth_users").
		Select("auth_users.user_id, auth_users.name").
		Joins("JOIN auth_roles ON auth_roles.role_id = auth_users.role_id").
		Where("UPPER(auth_roles.name) IN ?", []string{"ADMIN", "AGENT"}).
		Order("auth_users.name ASC, auth_users.user_id ASC").
		Find(&agents).Error
	if err != nil {
		return nil, err
	}

	return agents, nil
}

func dedupeUintIDs(ids []uint) []uint {
	seen := make(map[uint]struct{}, len(ids))
	result := make([]uint, 0, len(ids))

	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}

	return result
}
