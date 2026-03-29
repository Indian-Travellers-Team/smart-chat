package handlers

import (
	"net/http"
	"smart-chat/internal/constants"
	"time"

	"smart-chat/internal/services/analytics"

	"github.com/gin-gonic/gin"
)

func GetConversationsCountLast30DaysHandler(analyticsService *analytics.AnalyticsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		startDateStr := c.Query("startdate")
		endDateStr := c.Query("enddate")

		var (
			series []analytics.DailyConversationCount
			err    error
			days   int
		)

		if startDateStr == "" && endDateStr == "" {
			series, err = analyticsService.GetConversationCountsLastNDays(30, time.Now())
			days = 30
		} else {
			if startDateStr == "" || endDateStr == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrDateRangeRequired})
				return
			}

			startDate, parseErr := time.Parse(constants.DateFormat, startDateStr)
			if parseErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidStartDate})
				return
			}

			endDate, parseErr := time.Parse(constants.DateFormat, endDateStr)
			if parseErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidEndDate})
				return
			}

			if endDate.Before(startDate) {
				c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidDateRange})
				return
			}

			days = int(endDate.Sub(startDate).Hours()/24) + 1
			if days > 30 {
				c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrDateRangeExceedsLimit})
				return
			}

			series, err = analyticsService.GetConversationCountsByDateRange(startDate, endDate)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching analytics"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"days":   days,
			"series": series,
		})
	}
}

func GetDashboardConversationSummaryHandler(analyticsService *analytics.AnalyticsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		summary, err := analyticsService.GetDashboardConversationSummary(time.Now())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching dashboard analytics"})
			return
		}

		c.JSON(http.StatusOK, summary)
	}
}
