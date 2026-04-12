package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"smart-chat/config"
	"smart-chat/internal/handlers"
	"smart-chat/internal/models"
	authUserConversation "smart-chat/internal/services/auth_user_conversation"
	"smart-chat/internal/services/slack"
	"smart-chat/tests/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateAuthUserConversationHandler_UnauthorizedWithoutToken(t *testing.T) {
	db, teardown := utils.SetupTestDB()
	defer teardown()

	service := authUserConversation.NewService(db)
	router := gin.New()
	router.PATCH("/conversations/tracking", handlers.UpdateAuthUserConversationHandler(service, mockTokenValidator{userID: "zitadel-agent-track"}, nil))

	payload := map[string]any{"conversation_id": 1, "started": true}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPatch, "/conversations/tracking", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Authorization Required")
}

func TestUpdateAuthUserConversationHandler_ForbiddenForViewer(t *testing.T) {
	db, teardown := utils.SetupTestDB()
	defer teardown()

	viewer := setupAuthUserWithRole(t, db, "VIEWER", "zitadel-viewer-track", "Viewer Track")
	_, _, conv, _ := utils.SetupTestEntities(db)
	require.NoError(t, db.Create(&models.AuthUserConversation{AuthUserID: viewer.UserID, ConversationID: conv.ID}).Error)

	service := authUserConversation.NewService(db)
	router := gin.New()
	router.PATCH("/conversations/tracking", handlers.UpdateAuthUserConversationHandler(service, mockTokenValidator{userID: "zitadel-viewer-track"}, nil))

	payload := map[string]any{"conversation_id": conv.ID, "started": true}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPatch, "/conversations/tracking", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusForbidden, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "agent or admin role required")
}

func TestUpdateAuthUserConversationHandler_SuccessAndSlackNotification(t *testing.T) {
	db, teardown := utils.SetupTestDB()
	defer teardown()

	agent := setupAuthUserWithRole(t, db, "AGENT", "zitadel-agent-track", "Agent Track")
	_, _, conv, _ := utils.SetupTestEntities(db)
	require.NoError(t, db.Create(&models.AuthUserConversation{AuthUserID: agent.UserID, ConversationID: conv.ID}).Error)

	slackRequests := make(chan string, 1)
	slackServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var payload map[string]string
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		slackRequests <- payload["text"]
		w.WriteHeader(http.StatusOK)
	}))
	defer slackServer.Close()

	slackService := slack.NewSlackService(&config.Config{
		SlackNotificationURL: slackServer.URL,
		SlackAlertURL:        slackServer.URL,
	}, db)

	service := authUserConversation.NewService(db)
	router := gin.New()
	router.PATCH("/conversations/tracking", handlers.UpdateAuthUserConversationHandler(service, mockTokenValidator{userID: "zitadel-agent-track"}, slackService))

	payload := map[string]any{
		"conversation_id": conv.ID,
		"started":         true,
		"resolved":        true,
		"comments":        "Started review and resolved",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPatch, "/conversations/tracking", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "\"status\":\"updated\"")
	assert.Contains(t, recorder.Body.String(), "Started review and resolved")

	var link models.AuthUserConversation
	err := db.Where("auth_user_id = ? AND conversation_id = ?", agent.UserID, conv.ID).Take(&link).Error
	require.NoError(t, err)
	assert.True(t, link.Started)
	assert.True(t, link.Resolved)
	assert.Equal(t, "Started review and resolved", link.Comments)

	select {
	case msg := <-slackRequests:
		assert.Contains(t, msg, "conversation ID: *")
		assert.Contains(t, msg, "Agent Track")
		assert.Contains(t, msg, "started=true")
		assert.Contains(t, msg, "resolved=true")
	case <-time.After(2 * time.Second):
		t.Fatal("expected Slack notification")
	}
}