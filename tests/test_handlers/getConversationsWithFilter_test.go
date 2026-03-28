package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"smart-chat/internal/handlers"
	convHistory "smart-chat/internal/services/conversation_history"
	"smart-chat/tests/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type conversationsResponse struct {
	Conversations []struct {
		ID uint `json:"id"`
	} `json:"conversations"`
}

func TestGetConversationsWithFiltersHandler_DefaultSortAsc(t *testing.T) {
	db, teardown := utils.SetupTestDB()
	defer teardown()

	_, _, firstConv, _ := utils.SetupTestEntities(db)

	secondConv := firstConv
	secondConv.ID = 0
	secondConv.CreatedAt = time.Time{}
	if err := db.Create(&secondConv).Error; err != nil {
		t.Fatalf("failed to create second conversation: %v", err)
	}

	olderTime := time.Now().Add(-2 * time.Hour)
	newerTime := time.Now().Add(-1 * time.Hour)
	assert.NoError(t, db.Model(&firstConv).Update("created_at", olderTime).Error)
	assert.NoError(t, db.Model(&secondConv).Update("created_at", newerTime).Error)

	historyService := convHistory.NewConvHistoryService(db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/conversations", handlers.GetConversationsWithFiltersHandler(historyService))

	req, _ := http.NewRequest(http.MethodGet, "/conversations?page=1&limit=10", nil)
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

	secondConv := firstConv
	secondConv.ID = 0
	secondConv.CreatedAt = time.Time{}
	if err := db.Create(&secondConv).Error; err != nil {
		t.Fatalf("failed to create second conversation: %v", err)
	}

	olderTime := time.Now().Add(-2 * time.Hour)
	newerTime := time.Now().Add(-1 * time.Hour)
	assert.NoError(t, db.Model(&firstConv).Update("created_at", olderTime).Error)
	assert.NoError(t, db.Model(&secondConv).Update("created_at", newerTime).Error)

	historyService := convHistory.NewConvHistoryService(db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/conversations", handlers.GetConversationsWithFiltersHandler(historyService))

	req, _ := http.NewRequest(http.MethodGet, "/conversations?page=1&limit=10&sort=desc", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response conversationsResponse
	assert.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Len(t, response.Conversations, 2)
	assert.Equal(t, secondConv.ID, response.Conversations[0].ID)
	assert.Equal(t, firstConv.ID, response.Conversations[1].ID)
}
