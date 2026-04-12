package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"smart-chat/internal/authservice/zitadel"
	"smart-chat/internal/constants"
	"smart-chat/internal/handlers"
	"smart-chat/internal/models"
	authUserConversation "smart-chat/internal/services/auth_user_conversation"
	convHistory "smart-chat/internal/services/conversation_history"
	"smart-chat/tests/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type conversationsResponse struct {
	Conversations []struct {
		ID uint `json:"id"`
	} `json:"conversations"`
}

type mockTokenValidator struct {
	userID string
	err    error
}

func (m mockTokenValidator) ValidateToken(_ context.Context, _ string) (*zitadel.ValidateTokenUser, error) {
	if m.err != nil {
		return nil, m.err
	}
	uid := m.userID
	return &zitadel.ValidateTokenUser{ID: &uid}, nil
}

func setupAdminAuthUser(t *testing.T, db *gorm.DB, zitadelUserID string) {
	t.Helper()

	adminRole := models.AuthRole{Name: "ADMIN"}
	if err := db.Where("name = ?", adminRole.Name).FirstOrCreate(&adminRole).Error; err != nil {
		t.Fatalf("failed to create admin role: %v", err)
	}

	name := "Admin"
	email := zitadelUserID + "@example.com"
	authUser := models.AuthUser{
		ZitadelUserID: zitadelUserID,
		Name:          &name,
		Email:         &email,
		RoleID:        adminRole.RoleID,
	}
	if err := db.Create(&authUser).Error; err != nil {
		t.Fatalf("failed to create auth user: %v", err)
	}
}

func setupAuthUserWithRole(t *testing.T, db *gorm.DB, roleName, zitadelUserID, displayName string) models.AuthUser {
	t.Helper()

	role := models.AuthRole{Name: roleName}
	if err := db.Where("name = ?", role.Name).FirstOrCreate(&role).Error; err != nil {
		t.Fatalf("failed to create role %s: %v", roleName, err)
	}

	email := zitadelUserID + "@example.com"
	authUser := models.AuthUser{
		ZitadelUserID: zitadelUserID,
		Name:          &displayName,
		Email:         &email,
		RoleID:        role.RoleID,
	}
	if err := db.Create(&authUser).Error; err != nil {
		t.Fatalf("failed to create auth user %s: %v", zitadelUserID, err)
	}

	return authUser
}

func setupRouter(db *gorm.DB, zitadelUserID string) *gin.Engine {
	historyService := convHistory.NewConvHistoryService(db)
	authUserConversationService := authUserConversation.NewService(db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET(
		"/conversations",
		handlers.GetConversationsWithFiltersHandler(
			historyService,
			authUserConversationService,
			mockTokenValidator{userID: zitadelUserID},
		),
	)

	return router
}

func TestGetConversationsWithFiltersHandler_DefaultSortAsc(t *testing.T) {
	db, teardown := utils.SetupTestDB()
	defer teardown()

	_, _, firstConv, _ := utils.SetupTestEntities(db)
	setupAdminAuthUser(t, db, "zitadel-admin-default-asc")

	secondConv := firstConv
	secondConv.ID = 0
	secondConv.CreatedAt = time.Time{}
	assert.NoError(t, db.Create(&secondConv).Error)

	olderTime := time.Now().Add(-2 * time.Hour)
	newerTime := time.Now().Add(-1 * time.Hour)
	assert.NoError(t, db.Model(&firstConv).Update("created_at", olderTime).Error)
	assert.NoError(t, db.Model(&secondConv).Update("created_at", newerTime).Error)

	router := setupRouter(db, "zitadel-admin-default-asc")
	req, _ := http.NewRequest(http.MethodGet, "/conversations?page=1&limit=10", nil)
	req.Header.Set("Authorization", "Bearer admin-token")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response conversationsResponse
	assert.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Len(t, response.Conversations, 2)
	assert.Equal(t, firstConv.ID, response.Conversations[0].ID)
	assert.Equal(t, secondConv.ID, response.Conversations[1].ID)
}

func TestGetConversationsWithFiltersHandler_SortDesc(t *testing.T) {
	db, teardown := utils.SetupTestDB()
	defer teardown()

	_, _, firstConv, _ := utils.SetupTestEntities(db)
	setupAdminAuthUser(t, db, "zitadel-admin-sort-desc")

	secondConv := firstConv
	secondConv.ID = 0
	secondConv.CreatedAt = time.Time{}
	assert.NoError(t, db.Create(&secondConv).Error)

	olderTime := time.Now().Add(-2 * time.Hour)
	newerTime := time.Now().Add(-1 * time.Hour)
	assert.NoError(t, db.Model(&firstConv).Update("created_at", olderTime).Error)
	assert.NoError(t, db.Model(&secondConv).Update("created_at", newerTime).Error)

	router := setupRouter(db, "zitadel-admin-sort-desc")
	req, _ := http.NewRequest(http.MethodGet, "/conversations?page=1&limit=10&sort=desc", nil)
	req.Header.Set("Authorization", "Bearer admin-token")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response conversationsResponse
	assert.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Len(t, response.Conversations, 2)
	assert.Equal(t, secondConv.ID, response.Conversations[0].ID)
	assert.Equal(t, firstConv.ID, response.Conversations[1].ID)
}

func TestGetConversationsWithFiltersHandler_SourceFilter(t *testing.T) {
	db, teardown := utils.SetupTestDB()
	defer teardown()

	_, sessionOne, convOne, _ := utils.SetupTestEntities(db)
	assert.NoError(t, db.Model(&sessionOne).Update("source", "website").Error)
	_, sessionTwo, convTwo, _ := utils.SetupTestEntities(db)
	assert.NoError(t, db.Model(&sessionTwo).Update("source", "whatsapp").Error)
	setupAdminAuthUser(t, db, "zitadel-admin-source")

	router := setupRouter(db, "zitadel-admin-source")
	req, _ := http.NewRequest(http.MethodGet, "/conversations?page=1&limit=10&source=whatsapp", nil)
	req.Header.Set("Authorization", "Bearer admin-token")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response conversationsResponse
	assert.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Len(t, response.Conversations, 1)
	assert.Equal(t, convTwo.ID, response.Conversations[0].ID)
	assert.NotEqual(t, convOne.ID, response.Conversations[0].ID)
}

func TestGetConversationsWithFiltersHandler_ConversationIDFilter(t *testing.T) {
	db, teardown := utils.SetupTestDB()
	defer teardown()

	_, _, convOne, _ := utils.SetupTestEntities(db)
	_, _, convTwo, _ := utils.SetupTestEntities(db)
	setupAdminAuthUser(t, db, "zitadel-admin-conv-id")

	router := setupRouter(db, "zitadel-admin-conv-id")
	url := "/conversations?page=1&limit=10&conversationid=" + strconv.FormatUint(uint64(convOne.ID), 10)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer admin-token")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response conversationsResponse
	assert.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Len(t, response.Conversations, 1)
	assert.Equal(t, convOne.ID, response.Conversations[0].ID)
	assert.NotEqual(t, convTwo.ID, response.Conversations[0].ID)
}

func TestGetConversationsWithFiltersHandler_InvalidConversationID(t *testing.T) {
	db, teardown := utils.SetupTestDB()
	defer teardown()
	setupAdminAuthUser(t, db, "zitadel-admin-invalid-id")

	router := setupRouter(db, "zitadel-admin-invalid-id")
	req, _ := http.NewRequest(http.MethodGet, "/conversations?page=1&limit=10&conversationid=abc", nil)
	req.Header.Set("Authorization", "Bearer admin-token")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), constants.ErrInvalidConversationID)
}

func TestGetConversationsWithFiltersHandler_UnauthorizedWithoutToken(t *testing.T) {
	db, teardown := utils.SetupTestDB()
	defer teardown()

	_, _, _, _ = utils.SetupTestEntities(db)
	setupAdminAuthUser(t, db, "zitadel-admin-no-token")

	router := setupRouter(db, "zitadel-admin-no-token")
	req, _ := http.NewRequest(http.MethodGet, "/conversations?page=1&limit=10", nil)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Authorization Required")
}

func TestGetConversationsWithFiltersHandler_AgentGetsOnlyAssignedConversations(t *testing.T) {
	db, teardown := utils.SetupTestDB()
	defer teardown()

	_, _, convOne, _ := utils.SetupTestEntities(db)
	_, _, convTwo, _ := utils.SetupTestEntities(db)

	agentUser := setupAuthUserWithRole(t, db, "AGENT", "zitadel-agent-1", "Agent One")
	_ = setupAuthUserWithRole(t, db, "AGENT", "zitadel-agent-2", "Agent Two")

	assert.NoError(t, db.Create(&models.AuthUserConversation{
		AuthUserID:     agentUser.UserID,
		ConversationID: convOne.ID,
	}).Error)

	router := setupRouter(db, "zitadel-agent-1")
	req, _ := http.NewRequest(http.MethodGet, "/conversations?page=1&limit=10", nil)
	req.Header.Set("Authorization", "Bearer agent-token")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response conversationsResponse
	assert.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Len(t, response.Conversations, 1)
	assert.Equal(t, convOne.ID, response.Conversations[0].ID)
	assert.NotEqual(t, convTwo.ID, response.Conversations[0].ID)
}
