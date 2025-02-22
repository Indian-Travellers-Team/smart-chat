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

// New function to create user initial query by calling the external API
func createUserInitialQuery(toolCall openai.ToolCall, db *gorm.DB, conversationID uint, messageId uint) (string, error) {
	var args struct {
		NoOfPeople           int    `json:"no_of_people"`
		PreferredDestination string `json:"preferred_destination"`
		PreferredDate        string `json:"preferred_date"`
	}

	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return "", err
	}

	var conversation models.Conversation
	if err := db.Preload("Session.User").First(&conversation, conversationID).Error; err != nil {
		log.Printf("Error fetching conversation details: %v", err)
		return "", err
	}
	mobile := conversation.Session.User.Mobile

	// Call the external API to create the user initial query
	threadID := fmt.Sprintf("%v", conversationID)
	response, err := external.CreateUserInitialQuery(threadID, mobile, args.NoOfPeople, args.PreferredDestination, args.PreferredDate)
	if err != nil {
		log.Printf("Error calling external API: %v", err)
		return "", err
	}

	log.Printf("API response: %v", response.Message)
	return "Success", nil
}
