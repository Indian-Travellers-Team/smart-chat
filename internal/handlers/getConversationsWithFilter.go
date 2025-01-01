package handlers

import (
	"log"
	"net/http"
	convHistory "smart-chat/internal/services/conversation_history"
	"smart-chat/internal/services/conversation_history/specification"
	"time"

	"github.com/gin-gonic/gin"
)

// GetConversations fetches conversations based on query parameters for filtering.
func GetConversationsWithFiltersHandler(historyService *convHistory.ConvHistoryService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if filtering by dates
		startDateStr, endDateStr := c.Query("startdate"), c.Query("enddate")
		var specs []specification.Specification

		if startDateStr != "" && endDateStr != "" {
			startDate, err := time.Parse("02-01-2006", startDateStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
				return
			}

			endDate, err := time.Parse("02-01-2006", endDateStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
				return
			}

			specs = append(specs, specification.ByDateRange{StartDate: startDate, EndDate: endDate})
		}

		conversations, err := historyService.GetConversations(specs...)
		if err != nil {
			log.Printf("Error fetching conversations: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching conversations"})
			return
		}

		formattedConversations := []gin.H{}
		for _, conv := range conversations {
			formattedConversations = append(formattedConversations, gin.H{
				"id":        conv.ID,
				"createdAt": conv.CreatedAt.Format(time.RFC3339),
				"username":  conv.Session.User.Name,
				"mobile":    conv.Session.User.Mobile,
			})
		}

		c.JSON(http.StatusOK, gin.H{"conversations": formattedConversations})
	}
}
