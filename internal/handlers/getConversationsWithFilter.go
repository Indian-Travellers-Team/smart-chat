package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"smart-chat/internal/constants"
	convHistory "smart-chat/internal/services/conversation_history"
	"smart-chat/internal/services/conversation_history/specification"

	"github.com/gin-gonic/gin"
)

// GetConversationsWithFiltersHandler fetches conversations with optional filters and pagination.
func GetConversationsWithFiltersHandler(historyService *convHistory.ConvHistoryService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var specs []specification.Specification

		// 1. Handle optional date range filters.
		startDateStr, endDateStr := c.Query("startdate"), c.Query("enddate")
		if startDateStr != "" && endDateStr != "" {
			startDate, err := time.Parse(constants.DateFormat, startDateStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidStartDate})
				return
			}

			endDate, err := time.Parse(constants.DateFormat, endDateStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidEndDate})
				return
			}

			specs = append(specs, specification.ByDateRange{
				StartDate: startDate,
				EndDate:   endDate,
			})
		}

		// 2. Handle mobile filter.
		if mobile := c.Query("mobile"); mobile != "" {
			specs = append(specs, specification.ByMobile{Mobile: mobile})
		}

		// 3. Read pagination parameters (defaults: page=1, limit=20).
		pageStr := c.DefaultQuery("page", constants.DefaultPageStr)
		limitStr := c.DefaultQuery("limit", constants.DefaultLimitStr)

		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = constants.DefaultPage
		}

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			limit = constants.DefaultLimit
		}

		offset := (page - 1) * limit

		// 4. Fetch total count (for pagination metadata).
		total, err := historyService.CountConversations(specs...)
		if err != nil {
			log.Printf("Error counting conversations: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching conversations count"})
			return
		}

		// 5. Fetch the paginated conversations.
		conversations, err := historyService.GetConversations(offset, limit, specs...)
		if err != nil {
			log.Printf("Error fetching conversations: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching conversations"})
			return
		}

		// 6. Format the response.
		formattedConversations := make([]gin.H, 0, len(conversations))
		for _, conv := range conversations {
			formattedConversations = append(formattedConversations, gin.H{
				"id":        conv.ID,
				"createdAt": conv.CreatedAt.Format(time.RFC3339),
				"username":  conv.Session.User.Name,
				"mobile":    conv.Session.User.Mobile,
				"source":    conv.Session.Source,
			})
		}

		// 7. Return conversations plus pagination info.
		c.JSON(http.StatusOK, gin.H{
			"conversations": formattedConversations,
			"pagination": gin.H{
				"page":  page,
				"limit": limit,
				"total": total,
			},
		})
	}
}
