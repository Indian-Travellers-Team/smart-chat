package conversation

import (
	"encoding/json"
	"fmt"
	"log"
	"smart-chat/cache"
	"smart-chat/external"
	"smart-chat/internal/models"

	openai "github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
)

func handleGetPackageDetails(toolCall openai.ToolCall, db *gorm.DB, conversationID uint, messageId uint) (*external.PackageDetails, error) {
	var args struct {
		PackageID int `json:"package_id"`
	}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf(cache.CacheKeys.GetPackage.Key, args.PackageID)

	var packageDetails *external.PackageDetails
	// Try to get the details from cache
	err := cache.GetCache(cacheKey, &packageDetails)
	if err != nil {
		log.Println("Cache miss for package details, fetching from external source")
		fetchedDetails, fetchErr := external.GetPackageDetails(args.PackageID)
		if fetchErr != nil {
			return nil, fetchErr
		}

		// Update packageDetails with fetched details
		packageDetails = fetchedDetails

		// Cache the newly fetched details
		cacheErr := cache.SetCache(cacheKey, fetchedDetails, cache.CacheKeys.GetPackage.TTL)
		if cacheErr != nil {
			log.Printf("Error caching package details: %v", cacheErr)
		}
	} else {
		log.Println("Cache hit for package details")
	}

	// Save the function call in the database
	functionCall := models.FunctionCall{
		ConversationID: conversationID,
		MessageID:      messageId,
		Name:           toolCall.Function.Name,
		Args:           []byte(toolCall.Function.Arguments),
	}
	if createErr := db.Create(&functionCall).Error; createErr != nil {
		return nil, createErr
	}

	// Return the package details directly
	return packageDetails, nil
}
