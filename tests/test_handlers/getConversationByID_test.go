package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"smart-chat/internal/handlers"
	"smart-chat/internal/models"
	authUserConversation "smart-chat/internal/services/auth_user_conversation"
	convHistory "smart-chat/internal/services/conversation_history"
	"smart-chat/tests/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupConversationByIDRouter(db *gorm.DB, zitadelUserID string) *gin.Engine {
	historyService := convHistory.NewConvHistoryService(db)
	authUserConversationService := authUserConversation.NewService(db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET(
		"/conversation/:id",
		handlers.GetConversationByIDHandler(
			historyService,
			authUserConversationService,
			mockTokenValidator{userID: zitadelUserID},
		),
	)

	return router
}

func TestGetConversationByIDHandler_UnauthorizedWithoutToken(t *testing.T) {
	db, teardown := utils.SetupTestDB()
	defer teardown()

	_, _, conv, _ := utils.SetupTestEntities(db)
	setupAuthUserWithRole(t, db, "ADMIN", "zitadel-admin-conversation-no-token", "Admin Conversation")

	router := setupConversationByIDRouter(db, "zitadel-admin-conversation-no-token")
	req, _ := http.NewRequest(http.MethodGet, "/conversation/"+strconv.FormatUint(uint64(conv.ID), 10), nil)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Authorization Required")
}

func TestGetConversationByIDHandler_AllowsAdmin(t *testing.T) {
	db, teardown := utils.SetupTestDB()
	defer teardown()

	_, _, conv, _ := utils.SetupTestEntities(db)
	setupAuthUserWithRole(t, db, "ADMIN", "zitadel-admin-conversation", "Admin Conversation")

	router := setupConversationByIDRouter(db, "zitadel-admin-conversation")
	req, _ := http.NewRequest(http.MethodGet, "/conversation/"+strconv.FormatUint(uint64(conv.ID), 10), nil)
	req.Header.Set("Authorization", "Bearer admin-token")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Hello")
	assert.Contains(t, recorder.Body.String(), "Hi there!")
}

func TestGetConversationByIDHandler_AllowsAssignedAgent(t *testing.T) {
	db, teardown := utils.SetupTestDB()
	defer teardown()

	_, _, conv, _ := utils.SetupTestEntities(db)
	agent := setupAuthUserWithRole(t, db, "AGENT", "zitadel-agent-conversation", "Agent Conversation")

	assert.NoError(t, db.Create(&models.AuthUserConversation{
		AuthUserID:     agent.UserID,
		ConversationID: conv.ID,
	}).Error)

	router := setupConversationByIDRouter(db, "zitadel-agent-conversation")
	req, _ := http.NewRequest(http.MethodGet, "/conversation/"+strconv.FormatUint(uint64(conv.ID), 10), nil)
	req.Header.Set("Authorization", "Bearer agent-token")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Hello")
	assert.Contains(t, recorder.Body.String(), "Hi there!")
}

func TestGetConversationByIDHandler_HidesConversationFromUnassignedAgent(t *testing.T) {
	db, teardown := utils.SetupTestDB()
	defer teardown()

	_, _, conv, _ := utils.SetupTestEntities(db)
	_ = setupAuthUserWithRole(t, db, "AGENT", "zitadel-agent-unassigned", "Agent Unassigned")

	router := setupConversationByIDRouter(db, "zitadel-agent-unassigned")
	req, _ := http.NewRequest(http.MethodGet, "/conversation/"+strconv.FormatUint(uint64(conv.ID), 10), nil)
	req.Header.Set("Authorization", "Bearer agent-token")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Conversation not found")
}