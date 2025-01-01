package handlers_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"smart-chat/internal/handlers"
	middleware "smart-chat/internal/middlewares"
	"smart-chat/internal/services/conversation"
	"smart-chat/tests/utils"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetMessagesHandlerWithMiddlewareSimulation(t *testing.T) {
	db, teardown := utils.SetupTestDB()
	defer teardown()

	_, session, _, _ := utils.SetupTestEntities(db)

	conversationService := conversation.NewConversationService(db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/messages", middleware.AuthSessionMiddleware(db), handlers.GetConversationHandler(conversationService))

	req, _ := http.NewRequest(http.MethodGet, "/messages", nil)
	req.Header.Set("Authorization", session.AuthToken)
	w := httptest.NewRecorder()

	ginContext, _ := gin.CreateTestContext(w)
	ginContext.Request = req
	ginContext.Set("session", session)

	router.ServeHTTP(w, req)

	log.Printf("%v", w.Body.String())

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Hello")
	assert.Contains(t, w.Body.String(), "Hi there!")
}
