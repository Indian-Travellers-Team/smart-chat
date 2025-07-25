package conversation

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"smart-chat/cache"
	"smart-chat/config"
	"smart-chat/external/indian_travellers"
	"smart-chat/internal/llm_service"
	"smart-chat/internal/models"
	"smart-chat/internal/services/slack"

	openai "github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
)

type ConversationExecutor struct {
	db                *gorm.DB
	indian_travellers *indian_travellers.Client
	slackService      *slack.SlackService
}

func NewConversationExecutor(db *gorm.DB) *ConversationExecutor {

	return &ConversationExecutor{
		db:                db,
		indian_travellers: indian_travellers.NewClient(config.Load()),
		slackService:      slack.NewSlackService(config.Load(), db),
	}
}

func (ce *ConversationExecutor) Execute(conversationID uint, userInput string, messageType models.MessageType, conversationState *ConversationState, whatsapp bool) (string, error) {
	packages, err := ce.getPackageListFromCache()
	if err != nil {
		log.Printf("Error getting package list: %v", err)
		ce.slackService.SendSlackAlertAsync(fmt.Sprintf("Error getting package list: *%v* for conversation ID: *%d*", err, conversationID))
		return "", err
	}
	messages := ce.prepareMessages(conversationState.ConversationHistory, packages, userInput, whatsapp)
	conversationState.ConversationHistory = messages
	var botResponse string
	for {
		if conversationState.State == ConversationStateEnd {
			break
		}
		botResponse, err = ce.processInput(conversationID, userInput, conversationState, whatsapp)
		if err != nil {
			return "", err
		}
	}

	return botResponse, nil
}

func (ce *ConversationExecutor) processInput(conversationID uint, userInput string, conversationState *ConversationState, whatsapp bool) (string, error) {
	var botResponse string
	var totalTokens uint
	var responseType models.MessageType
	var responseContent interface{}
	var err error
	if whatsapp {
		responseType, responseContent, totalTokens, err = llm_service.GetOpenAIResponsev2Whatsapp(conversationState.ConversationHistory)
		if err != nil {
			log.Printf("Error processing user input with OpenAI: %v", err)
			ce.slackService.SendSlackAlertAsync(fmt.Sprintf("Error processing user input with OpenAI: *%v* for conversation ID: *%d*", err, conversationID))
			return "", err
		}
	} else {
		responseType, responseContent, totalTokens, err = llm_service.GetOpenAIResponsev2(conversationState.ConversationHistory)
		if err != nil {
			log.Printf("Error processing user input with OpenAI: %v", err)
			ce.slackService.SendSlackAlertAsync(fmt.Sprintf("Error processing user input with OpenAI: *%v* for conversation ID: *%d*", err, conversationID))
			return "", err
		}
	}

	switch responseType {
	case models.MessageTypeFunctionCall:
		toolCall, ok := responseContent.(openai.ToolCall)
		if !ok {
			ce.slackService.SendSlackAlertAsync(fmt.Sprintf("Error asserting tool call from response content: *%v* for conversation ID: *%d*", err, conversationID))
			return "", errors.New("error asserting tool call from response content")
		}
		messageId, err := ce.updateConversation(conversationID, "", "", totalTokens, responseType)
		if err != nil {
			ce.slackService.SendSlackAlertAsync(fmt.Sprintf("Error updating conversation: *%v* for conversation ID: *%d*", err, conversationID))
			return "", err
		}
		functionResponse, err := processFunctionResponse(ce.indian_travellers, toolCall, ce.db, conversationID, messageId)
		if err != nil {
			log.Printf("Error processing function response: %v", err)
			ce.slackService.SendSlackAlertAsync(fmt.Sprintf("Error processing function response: *%v* for conversation ID: *%d*", err, conversationID))
			conversationState.EndState()
			return "we encountered an error while processing your request. Please try again later.", nil
		}
		functionResponseString, _ := json.Marshal(functionResponse)
		functionMessage := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleFunction,
			Content: string(functionResponseString),
			Name:    toolCall.Function.Name,
		}
		conversationState.AddToHistory(functionMessage)
	default:
		botResponse, _ = responseContent.(string)
		if _, err := ce.updateConversation(conversationID, userInput, botResponse, totalTokens, responseType); err != nil {
			ce.slackService.SendSlackAlertAsync(fmt.Sprintf("Error updating conversation: *%v* for conversation ID: *%d*", err, conversationID))
			return "", err
		}
	}
	conversationState.NextState(responseType)
	return botResponse, nil
}

func (ce *ConversationExecutor) prepareMessages(history []openai.ChatCompletionMessage, packages []indian_travellers.Package, userInput string, whatsapp bool) []openai.ChatCompletionMessage {
	var systemTemplate string
	if whatsapp {
		systemTemplate = llm_service.SystemMessageTemplateForWhatsapp(ce.indian_travellers, packages, 1)
	} else {
		systemTemplate = llm_service.SystemMessageTemplate(packages)
	}
	messages := append([]openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleSystem, Content: systemTemplate}}, history...)
	messages = append(messages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: userInput})
	return messages
}

func (ce *ConversationExecutor) getPackageListFromCache() ([]indian_travellers.Package, error) {
	cacheKey := fmt.Sprintf(cache.CacheKeys.GetPackageList.Key)
	var packages []indian_travellers.Package

	if err := cache.GetCache(cacheKey, &packages); err == nil {
		log.Println("Cache hit for package list")
		return packages, nil
	}

	log.Println("Cache miss for package list, fetching from external source")
	packages, err := ce.indian_travellers.GetPackageList()
	if err != nil {
		return nil, err
	}

	if err := cache.SetCache(cacheKey, packages, cache.CacheKeys.GetPackageList.TTL); err != nil {
		log.Printf("Error caching package list: %v", err)
	}

	return packages, nil
}

func (ce *ConversationExecutor) updateConversation(conversationID uint, userInput, botResponse string, totalTokens uint, messageType models.MessageType) (uint, error) {
	var visible bool
	switch messageType {
	case models.MessageTypeUserFix, models.MessageTypeOffTopic, models.MessageTypeFunctionCall:
		visible = false
	default:
		visible = true
	}

	messagePair := models.MessagePair{
		ConversationID: conversationID,
		User:           userInput,
		Bot:            botResponse,
		TotalTokens:    uint(totalTokens),
		Visible:        visible,
		Type:           messageType,
	}

	// Save the message pair to the database.
	if err := ce.db.Create(&messagePair).Error; err != nil {
		log.Printf("Error saving message pair: %v", err)
		return 0, err // Return 0 as the ID in case of error
	}

	return messagePair.ID, nil // Return the ID of the newly created message pair
}

func processFunctionResponse(indian_travellers_client *indian_travellers.Client, toolCall openai.ToolCall, db *gorm.DB, conversationID uint, messageId uint) (interface{}, error) {
	// Generalize the handling of function calls based on the function's name
	switch toolCall.Function.Name {
	case "get_package_details":
		return handleGetPackageDetails(indian_travellers_client, toolCall, db, conversationID, messageId)
	case "create_user_initial_query":
		return createUserInitialQuery(indian_travellers_client, toolCall, db, conversationID, messageId)
	case "create_user_final_booking":
		return createUserFinalBooking(indian_travellers_client, toolCall, db, conversationID, messageId)
	case "fetch_upcoming_trips":
		return fetchUpcomingTrips(indian_travellers_client, toolCall, db, conversationID, messageId)
	// Add more cases as necessary for other function types
	default:
		log.Printf("Unhandled function call: %s", toolCall.Function.Name)
		return nil, nil
	}
}
