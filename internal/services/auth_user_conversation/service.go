package auth_user_conversation

import (
	"errors"
	"fmt"
	"strings"
	"time"

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

type AuthPrincipal struct {
	UserID   uint
	RoleName string
}

type UpdateConversationTrackingInput struct {
	AuthUserID     uint
	ConversationID uint
	Started        *bool
	Resolved       *bool
	Comments       *string
}

type ConversationTracking struct {
	AuthUserID     uint
	ConversationID uint
	Started        bool
	Resolved       bool
	Comments       string
	AgentName      *string
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

func (s *Service) UpdateConversationTracking(input UpdateConversationTrackingInput) (*ConversationTracking, error) {
	if input.AuthUserID == 0 {
		return nil, errors.New("user_id is required")
	}
	if input.ConversationID == 0 {
		return nil, errors.New("conversation_id is required")
	}

	updates := make(map[string]any)
	if input.Started != nil {
		updates["started"] = *input.Started
	}
	if input.Resolved != nil {
		updates["resolved"] = *input.Resolved
	}
	if input.Comments != nil {
		updates["comments"] = strings.TrimSpace(*input.Comments)
	}
	if len(updates) == 0 {
		return nil, errors.New("at least one field must be provided for update")
	}

	result := s.db.Model(&models.AuthUserConversation{}).
		Where("auth_user_id = ? AND conversation_id = ?", input.AuthUserID, input.ConversationID).
		Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	type trackingRow struct {
		AuthUserID     uint    `gorm:"column:auth_user_id"`
		ConversationID uint    `gorm:"column:conversation_id"`
		Started        bool    `gorm:"column:started"`
		Resolved       bool    `gorm:"column:resolved"`
		Comments       string  `gorm:"column:comments"`
		AgentName      *string `gorm:"column:name"`
	}

	var row trackingRow
	err := s.db.
		Table("auth_user_conversation").
		Select("auth_user_conversation.auth_user_id, auth_user_conversation.conversation_id, auth_user_conversation.started, auth_user_conversation.resolved, auth_user_conversation.comments, auth_users.name").
		Joins("JOIN auth_users ON auth_users.user_id = auth_user_conversation.auth_user_id").
		Where("auth_user_conversation.auth_user_id = ? AND auth_user_conversation.conversation_id = ?", input.AuthUserID, input.ConversationID).
		Take(&row).Error
	if err != nil {
		return nil, err
	}

	return &ConversationTracking{
		AuthUserID:     row.AuthUserID,
		ConversationID: row.ConversationID,
		Started:        row.Started,
		Resolved:       row.Resolved,
		Comments:       row.Comments,
		AgentName:      row.AgentName,
	}, nil
}

func (s *Service) IsAdminByZitadelUserID(zitadelUserID string) (bool, error) {
	principal, err := s.GetAuthPrincipalByZitadelUserID(zitadelUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	return strings.EqualFold(strings.TrimSpace(principal.RoleName), "ADMIN"), nil
}

func (s *Service) GetAuthPrincipalByZitadelUserID(zitadelUserID string) (*AuthPrincipal, error) {
	zitadelUserID = strings.TrimSpace(zitadelUserID)
	if zitadelUserID == "" {
		return nil, errors.New("zitadel user id is required")
	}

	type principalLookup struct {
		UserID   uint   `gorm:"column:user_id"`
		RoleName string `gorm:"column:name"`
	}

	var row principalLookup
	err := s.db.
		Table("auth_users").
		Select("auth_users.user_id, auth_roles.name").
		Joins("JOIN auth_roles ON auth_roles.role_id = auth_users.role_id").
		Where("auth_users.zitadel_user_id = ?", zitadelUserID).
		Take(&row).Error
	if err != nil {
		return nil, err
	}

	return &AuthPrincipal{UserID: row.UserID, RoleName: row.RoleName}, nil
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

func (s *Service) GetAssignedAgentsByConversationIDs(conversationIDs []uint) (map[uint]*AgentUser, error) {
	result := make(map[uint]*AgentUser)
	if len(conversationIDs) == 0 {
		return result, nil
	}

	type assignedAgentRow struct {
		ConversationID uint      `gorm:"column:conversation_id"`
		UserID         uint      `gorm:"column:user_id"`
		Name           *string   `gorm:"column:name"`
		CreatedAt      time.Time `gorm:"column:created_at"`
	}

	rows := make([]assignedAgentRow, 0)
	err := s.db.
		Table("auth_user_conversation").
		Select("auth_user_conversation.conversation_id, auth_users.user_id, auth_users.name, auth_user_conversation.created_at").
		Joins("JOIN auth_users ON auth_users.user_id = auth_user_conversation.auth_user_id").
		Where("auth_user_conversation.conversation_id IN ?", conversationIDs).
		Order("auth_user_conversation.conversation_id ASC, auth_user_conversation.created_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		if _, exists := result[row.ConversationID]; exists {
			continue
		}
		result[row.ConversationID] = &AgentUser{UserID: row.UserID, Name: row.Name}
	}

	return result, nil
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
