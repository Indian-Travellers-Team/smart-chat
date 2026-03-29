package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"smart-chat/internal/constants"
	"testing"
	"time"

	"smart-chat/internal/handlers"
	"smart-chat/internal/models"
	"smart-chat/internal/services/analytics"
	"smart-chat/tests/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type analyticsResponse struct {
	Days   int `json:"days"`
	Series []struct {
		Date  string `json:"date"`
		Count int64  `json:"count"`
	} `json:"series"`
}

func TestGetConversationsCountLast30DaysHandler(t *testing.T) {
	db, teardown := utils.SetupTestDB()
	defer teardown()

	_, session, baseConversation, _ := utils.SetupTestEntities(db)

	assert.NoError(t, db.Model(&baseConversation).Update("created_at", time.Now().AddDate(0, 0, -45)).Error)

	conversationOne := models.Conversation{SessionID: session.ID}
	assert.NoError(t, db.Create(&conversationOne).Error)
	conversationOneTime := time.Now().AddDate(0, 0, -1)
	assert.NoError(t, db.Model(&conversationOne).Update("created_at", conversationOneTime).Error)

	conversationTwo := models.Conversation{SessionID: session.ID}
	assert.NoError(t, db.Create(&conversationTwo).Error)
	conversationTwoTime := time.Now().AddDate(0, 0, -10)
	assert.NoError(t, db.Model(&conversationTwo).Update("created_at", conversationTwoTime).Error)

	conversationOutsideRange := models.Conversation{SessionID: session.ID}
	assert.NoError(t, db.Create(&conversationOutsideRange).Error)
	assert.NoError(t, db.Model(&conversationOutsideRange).Update("created_at", time.Now().AddDate(0, 0, -35)).Error)

	analyticsService := analytics.NewAnalyticsService(db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/analytics/conversations/last-30-days", handlers.GetConversationsCountLast30DaysHandler(analyticsService))

	req, _ := http.NewRequest(http.MethodGet, "/analytics/conversations/last-30-days", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response analyticsResponse
	assert.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Equal(t, 30, response.Days)
	assert.Len(t, response.Series, 30)

	countByDate := make(map[string]int64, len(response.Series))
	var totalCount int64
	for _, item := range response.Series {
		countByDate[item.Date] = item.Count
		totalCount += item.Count
	}

	assert.Equal(t, int64(1), countByDate[conversationOneTime.UTC().Format("2006-01-02")])
	assert.Equal(t, int64(1), countByDate[conversationTwoTime.UTC().Format("2006-01-02")])
	assert.Equal(t, int64(2), totalCount)
}

func TestGetConversationsCountLast30DaysHandler_CustomRange(t *testing.T) {
	db, teardown := utils.SetupTestDB()
	defer teardown()

	_, session, _, _ := utils.SetupTestEntities(db)

	insideRange := models.Conversation{SessionID: session.ID}
	assert.NoError(t, db.Create(&insideRange).Error)
	insideRangeTime := time.Now().AddDate(0, 0, -5)
	assert.NoError(t, db.Model(&insideRange).Update("created_at", insideRangeTime).Error)

	outsideRange := models.Conversation{SessionID: session.ID}
	assert.NoError(t, db.Create(&outsideRange).Error)
	outsideRangeTime := time.Now().AddDate(0, 0, -20)
	assert.NoError(t, db.Model(&outsideRange).Update("created_at", outsideRangeTime).Error)

	analyticsService := analytics.NewAnalyticsService(db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/analytics/conversations/last-30-days", handlers.GetConversationsCountLast30DaysHandler(analyticsService))

	startDate := time.Now().AddDate(0, 0, -10).Format(constants.DateFormat)
	endDate := time.Now().AddDate(0, 0, -1).Format(constants.DateFormat)
	url := "/analytics/conversations/last-30-days?startdate=" + startDate + "&enddate=" + endDate

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response analyticsResponse
	assert.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Equal(t, 10, response.Days)
	assert.Len(t, response.Series, 10)

	countByDate := make(map[string]int64, len(response.Series))
	for _, item := range response.Series {
		countByDate[item.Date] = item.Count
	}

	assert.Equal(t, int64(1), countByDate[insideRangeTime.UTC().Format("2006-01-02")])
	assert.Equal(t, int64(0), countByDate[outsideRangeTime.UTC().Format("2006-01-02")])
}

func TestGetConversationsCountLast30DaysHandler_MissingRangeParam(t *testing.T) {
	db, teardown := utils.SetupTestDB()
	defer teardown()

	analyticsService := analytics.NewAnalyticsService(db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/analytics/conversations/last-30-days", handlers.GetConversationsCountLast30DaysHandler(analyticsService))

	req, _ := http.NewRequest(http.MethodGet, "/analytics/conversations/last-30-days?startdate=01-03-2026", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), constants.ErrDateRangeRequired)
}

func TestGetConversationsCountLast30DaysHandler_RangeExceeds30Days(t *testing.T) {
	db, teardown := utils.SetupTestDB()
	defer teardown()

	analyticsService := analytics.NewAnalyticsService(db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/analytics/conversations/last-30-days", handlers.GetConversationsCountLast30DaysHandler(analyticsService))

	req, _ := http.NewRequest(http.MethodGet, "/analytics/conversations/last-30-days?startdate=01-01-2026&enddate=05-02-2026", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), constants.ErrDateRangeExceedsLimit)
}
