package handlers

import (
	"net/http"
	"strconv"

	userService "smart-chat/internal/services/user"

	"github.com/gin-gonic/gin"
)

// ClientUserDetailsHandler fetches user details based on a conversation ID provided as a query parameter.
func ClientUserDetailsHandler(us *userService.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		convIDStr := c.Query("conv_id")
		if convIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "conv_id query parameter is required"})
			return
		}

		convID, err := strconv.Atoi(convIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "conv_id must be a valid integer"})
			return
		}

		details, err := us.GetUserDetailsByConversationID(uint(convID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user details"})
			return
		}

		c.JSON(http.StatusOK, details)
	}
}
