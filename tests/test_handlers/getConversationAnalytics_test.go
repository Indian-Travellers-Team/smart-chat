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
		Date            string `json:"date"`
		Count           int64  `json:"count"`
		ConversationIDs []uint `json:"conversationIds"`
	} `json:"series"`
}

type dashboardConversationSummaryResponse struct {
	TotalConversations        int64   `json:"totalConversations"`
	CurrentWeekConversations  int64   `json:"currentWeekConversations"`
	PreviousWeekConversations int64   `json:"previousWeekConversations"`
	WeeklyChangePercentage    float64 `json:"weeklyChangePercentage"`
	Trend                     string  `json:"trend"`
	PeriodLabel               string  `json:"periodLabel"`
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
	conversationIDsByDate := make(map[string][]uint, len(response.Series))
	var totalCount int64
	for _, item := range response.Series {
		countByDate[item.Date] = item.Count
		conversationIDsByDate[item.Date] = item.ConversationIDs
		totalCount += item.Count
	}

	assert.Equal(t, int64(1), countByDate[conversationOneTime.UTC().Format("2006-01-02")])
	assert.Equal(t, int64(1), countByDate[conversationTwoTime.UTC().Format("2006-01-02")])
	assert.Contains(t, conversationIDsByDate[conversationOneTime.UTC().Format("2006-01-02")], conversationOne.ID)
	assert.Contains(t, conversationIDsByDate[conversationTwoTime.UTC().Format("2006-01-02")], conversationTwo.ID)
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
	conversationIDsByDate := make(map[string][]uint, len(response.Series))
	for _, item := range response.Series {
		countByDate[item.Date] = item.Count
		conversationIDsByDate[item.Date] = item.ConversationIDs
	}

	assert.Equal(t, int64(1), countByDate[insideRangeTime.UTC().Format("2006-01-02")])
	assert.Equal(t, int64(0), countByDate[outsideRangeTime.UTC().Format("2006-01-02")])
	assert.Contains(t, conversationIDsByDate[insideRangeTime.UTC().Format("2006-01-02")], insideRange.ID)
	assert.Empty(t, conversationIDsByDate[outsideRangeTime.UTC().Format("2006-01-02")])
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

func TestGetDashboardConversationSummaryHandler(t *testing.T) {
	db, teardown := utils.SetupTestDB()
	defer teardown()

	_, session, baseConversation, _ := utils.SetupTestEntities(db)
	assert.NoError(t, db.Model(&baseConversation).Update("created_at", time.Now().AddDate(0, 0, -40)).Error)

	now := time.Now()
	currentWeekStart := testStartOfWeek(now)
	previousWeekStart := currentWeekStart.AddDate(0, 0, -7)

	currentWeekConversationOne := models.Conversation{SessionID: session.ID}
	assert.NoError(t, db.Create(&currentWeekConversationOne).Error)
	assert.NoError(t, db.Model(&currentWeekConversationOne).Update("created_at", currentWeekStart.Add(2*time.Hour)).Error)

	currentWeekConversationTwo := models.Conversation{SessionID: session.ID}
	assert.NoError(t, db.Create(&currentWeekConversationTwo).Error)
	assert.NoError(t, db.Model(&currentWeekConversationTwo).Update("created_at", currentWeekStart.AddDate(0, 0, 1).Add(3*time.Hour)).Error)

	previousWeekConversation := models.Conversation{SessionID: session.ID}
	assert.NoError(t, db.Create(&previousWeekConversation).Error)
	assert.NoError(t, db.Model(&previousWeekConversation).Update("created_at", previousWeekStart.Add(4*time.Hour)).Error)

	analyticsService := analytics.NewAnalyticsService(db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/analytics/dashboard/conversations-summary", handlers.GetDashboardConversationSummaryHandler(analyticsService))

	req, _ := http.NewRequest(http.MethodGet, "/analytics/dashboard/conversations-summary", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response dashboardConversationSummaryResponse
	assert.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Equal(t, int64(4), response.TotalConversations)
	assert.Equal(t, int64(2), response.CurrentWeekConversations)
	assert.Equal(t, int64(1), response.PreviousWeekConversations)
	assert.Equal(t, 100.0, response.WeeklyChangePercentage)
	assert.Equal(t, "up", response.Trend)
	assert.Equal(t, "this week", response.PeriodLabel)
}

func testStartOfWeek(now time.Time) time.Time {
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return start.AddDate(0, 0, -(weekday - 1))
}
