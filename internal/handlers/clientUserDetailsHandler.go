package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"smart-chat/cache"
	userService "smart-chat/internal/services/user"

	"github.com/gin-gonic/gin"
)

// ClientUserDetailsHandler fetches user details based on a conversation ID provided as a query parameter.
// It first checks the cache, and if not found, calls the user service and then caches the result.
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

		// Build the cache key using the format defined in cache.CacheKeys.UserDetails.Key
		cacheKey := fmt.Sprintf(cache.CacheKeys.UserDetails.Key, convID)

		// Try to get the user details from the cache
		var details userService.UserDetails
		if err := cache.GetCache(cacheKey, &details); err == nil {
			c.JSON(http.StatusOK, details)
			return
		}

		// If cache miss, fetch user details from the service
		details, err = us.GetUserDetailsByConversationID(uint(convID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user details"})
			return
		}

		// Set the fetched details in cache with the TTL defined in cache.CacheKeys.UserDetails.TTL
		if err := cache.SetCache(cacheKey, details, cache.CacheKeys.UserDetails.TTL); err != nil {
			log.Println("Unable to set cache for user details")
		}

		c.JSON(http.StatusOK, details)
	}
}
