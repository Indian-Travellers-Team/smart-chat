package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"smart-chat/internal/handlers"
	authUserConversation "smart-chat/internal/services/auth_user_conversation"
	"smart-chat/tests/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type agentsResponse struct {
	Agents []struct {
		UserID uint    `json:"user_id"`
		Name   *string `json:"name"`
	} `json:"agents"`
}

func TestGetAgentsHandler_UnauthorizedWithoutToken(t *testing.T) {
	db, teardown := utils.SetupTestDB()
	defer teardown()

	service := authUserConversation.NewService(db)
	router := gin.New()
	router.GET("/agents", handlers.GetAgentsHandler(service, mockTokenValidator{userID: "zitadel-admin"}))

	req, _ := http.NewRequest(http.MethodGet, "/agents", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Authorization Required")
}

func TestGetAgentsHandler_ForbiddenForNonAdmin(t *testing.T) {
	db, teardown := utils.SetupTestDB()
	defer teardown()

	_ = setupAuthUserWithRole(t, db, "AGENT", "zitadel-agent-only", "Agent Only")
	_ = setupAuthUserWithRole(t, db, "AGENT", "zitadel-agent-other", "Agent Other")

	service := authUserConversation.NewService(db)
	router := gin.New()
	router.GET("/agents", handlers.GetAgentsHandler(service, mockTokenValidator{userID: "zitadel-agent-only"}))

	req, _ := http.NewRequest(http.MethodGet, "/agents", nil)
	req.Header.Set("Authorization", "Bearer token")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusForbidden, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "admin role required")
}

func TestGetAgentsHandler_SuccessForAdmin(t *testing.T) {
	db, teardown := utils.SetupTestDB()
	defer teardown()

	_ = setupAuthUserWithRole(t, db, "ADMIN", "zitadel-admin-list", "Main Admin")
	_ = setupAuthUserWithRole(t, db, "AGENT", "zitadel-agent-list", "Assigned Agent")
	_ = setupAuthUserWithRole(t, db, "VIEWER", "zitadel-viewer-list", "Viewer User")

	service := authUserConversation.NewService(db)
	router := gin.New()
	router.GET("/agents", handlers.GetAgentsHandler(service, mockTokenValidator{userID: "zitadel-admin-list"}))

	req, _ := http.NewRequest(http.MethodGet, "/agents", nil)
	req.Header.Set("Authorization", "Bearer token")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response agentsResponse
	assert.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Len(t, response.Agents, 2)
}
