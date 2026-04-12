package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"smart-chat/internal/handlers"
	"smart-chat/internal/models"
	authUserConversation "smart-chat/internal/services/auth_user_conversation"
	"smart-chat/tests/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLinkAuthUserConversationsHandler_UnauthorizedWithoutToken(t *testing.T) {
	db, teardown := utils.SetupTestDB()
	defer teardown()

	service := authUserConversation.NewService(db)
	router := gin.New()
	router.POST("/conversations/link", handlers.LinkAuthUserConversationsHandler(service, mockTokenValidator{userID: "zitadel-admin"}))

	payload := map[string]any{"user_id": 1, "conversation_ids": []uint{1}}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "/conversations/link", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Authorization Required")
}

func TestLinkAuthUserConversationsHandler_ForbiddenForNonAdmin(t *testing.T) {
	db, teardown := utils.SetupTestDB()
	defer teardown()

	agent := setupAuthUserWithRole(t, db, "AGENT", "zitadel-agent-link", "Agent Link")
	_, _, conv, _ := utils.SetupTestEntities(db)

	service := authUserConversation.NewService(db)
	router := gin.New()
	router.POST("/conversations/link", handlers.LinkAuthUserConversationsHandler(service, mockTokenValidator{userID: "zitadel-agent-link"}))

	payload := map[string]any{"user_id": agent.UserID, "conversation_ids": []uint{conv.ID}}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "/conversations/link", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusForbidden, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "admin role required")
}

func TestLinkAuthUserConversationsHandler_SuccessForAdmin(t *testing.T) {
	db, teardown := utils.SetupTestDB()
	defer teardown()

	admin := setupAuthUserWithRole(t, db, "ADMIN", "zitadel-admin-link", "Admin Link")
	targetAgent := setupAuthUserWithRole(t, db, "AGENT", "zitadel-target-agent", "Target Agent")
	_, _, conv, _ := utils.SetupTestEntities(db)

	service := authUserConversation.NewService(db)
	router := gin.New()
	router.POST("/conversations/link", handlers.LinkAuthUserConversationsHandler(service, mockTokenValidator{userID: "zitadel-admin-link"}))

	payload := map[string]any{"user_id": targetAgent.UserID, "conversation_ids": []uint{conv.ID}}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "/conversations/link", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "\"status\":\"linked\"")
	assert.Contains(t, recorder.Body.String(), "\"linked_count\":1")

	var linkCount int64
	err := db.Model(&models.AuthUserConversation{}).
		Where("auth_user_id = ? AND conversation_id = ?", targetAgent.UserID, conv.ID).
		Count(&linkCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), linkCount)

	_ = admin
}
