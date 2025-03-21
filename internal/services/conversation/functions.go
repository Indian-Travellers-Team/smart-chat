package conversation

import (
	"encoding/json"
	"fmt"
	"log"
	"smart-chat/cache"
	external "smart-chat/external/indian_travellers"
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
		FunctionResponse: func() string {
			response, _ := json.Marshal(packageDetails)
			return string(response)
		}(),
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
	_, err := external.CreateUserInitialQuery(threadID, mobile, args.NoOfPeople, args.PreferredDestination, args.PreferredDate)
	if err != nil {
		log.Printf("Error calling external API: %v", err)
		return "", err
	}

	// Prepare message for LLM response
	llmMessage := fmt.Sprintf("The query has been created for %d people, with preferred destination: %s, and preferred date: %s",
		args.NoOfPeople, args.PreferredDestination, args.PreferredDate)

	// Save the function call in the database
	functionCall := models.FunctionCall{
		ConversationID: conversationID,
		MessageID:      messageId,
		Name:           toolCall.Function.Name,
		Args:           []byte(toolCall.Function.Arguments),
		FunctionResponse: func() string {
			response, _ := json.Marshal(llmMessage)
			return string(response)
		}(),
	}
	if createErr := db.Create(&functionCall).Error; createErr != nil {
		return "error", createErr
	}
	// Return the formatted message in JSON format
	return fmt.Sprintf(`{"ResponseToLLM": "%s"}`, llmMessage), nil
}

// New function to create user final booking by calling the external API
func createUserFinalBooking(toolCall openai.ToolCall, db *gorm.DB, conversationID uint, messageId uint) (string, error) {
	var args struct {
		TripID int `json:"trip_id"`
	}

	// Unmarshal the function arguments from the tool call
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return "", err
	}

	// Fetch mobile from the conversation session
	var conversation models.Conversation
	if err := db.Preload("Session.User").First(&conversation, conversationID).Error; err != nil {
		log.Printf("Error fetching conversation details: %v", err)
		return "", err
	}

	// Call the external API to create the final booking
	threadID := fmt.Sprintf("%v", conversationID)
	_, err := external.CreateUserFinalBooking(threadID, args.TripID)
	if err != nil {
		log.Printf("Error calling external API: %v", err)
		return "", err
	}

	// Prepare message for LLM response
	llmMessage := fmt.Sprintf("The final booking has been created for Trip ID: %d", args.TripID)

	// Save the function call in the database
	functionCall := models.FunctionCall{
		ConversationID: conversationID,
		MessageID:      messageId,
		Name:           toolCall.Function.Name,
		Args:           []byte(toolCall.Function.Arguments),
		FunctionResponse: func() string {
			response, _ := json.Marshal(llmMessage)
			return string(response)
		}(),
	}
	if createErr := db.Create(&functionCall).Error; createErr != nil {
		return "success", createErr
	}

	// Return the formatted message in JSON format
	return fmt.Sprintf(`{"ResponseToLLM": "%s"}`, llmMessage), nil
}

// New function to fetch upcoming trips for a given package ID
func fetchUpcomingTrips(toolCall openai.ToolCall, db *gorm.DB, conversationID uint, messageId uint) (*external.UpcomingTripsResponseInternal, error) {
	var args struct {
		PackageID int `json:"package_id"`
	}

	// Unmarshal the function arguments from the tool call
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return nil, err
	}

	// Fetch the upcoming trips from the external service
	upcomingTrips, err := external.GetUpcomingTrips(args.PackageID)
	if err != nil {
		log.Printf("Error fetching upcoming trips: %v", err)
		return nil, err
	}

	// Save the function call in the database
	functionCall := models.FunctionCall{
		ConversationID: conversationID,
		MessageID:      messageId,
		Name:           toolCall.Function.Name,
		Args:           []byte(toolCall.Function.Arguments),
		FunctionResponse: func() string {
			response, _ := json.Marshal(upcomingTrips)
			return string(response)
		}(),
	}
	if createErr := db.Create(&functionCall).Error; createErr != nil {
		return nil, createErr
	}

	// Return the upcoming trips response
	return upcomingTrips, nil
}
