package conversation

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"smart-chat/cache"
	"smart-chat/external"
	"smart-chat/internal/llm_service"
	"smart-chat/internal/models"

	openai "github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
)

type ConversationExecutor struct {
	db *gorm.DB
}

func NewConversationExecutor(db *gorm.DB) *ConversationExecutor {

	return &ConversationExecutor{
		db: db,
	}
}

func (ce *ConversationExecutor) Execute(conversationID uint, userInput string, messageType models.MessageType, conversationState *ConversationState, whatsapp bool) (string, error) {
	packages, err := ce.getPackageListFromCache()
	if err != nil {
		log.Printf("Error getting package list: %v", err)
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
			return "", err
		}
	} else {
		responseType, responseContent, totalTokens, err = llm_service.GetOpenAIResponsev2(conversationState.ConversationHistory)
		if err != nil {
			log.Printf("Error processing user input with OpenAI: %v", err)
			return "", err
		}
	}

	switch responseType {
	case models.MessageTypeFunctionCall:
		toolCall, ok := responseContent.(openai.ToolCall)
		if !ok {
			return "", errors.New("error asserting tool call from response content")
		}
		messageId, err := ce.updateConversation(conversationID, "", "", totalTokens, responseType)
		if err != nil {
			return "", err
		}
		functionResponse, err := processFunctionResponse(toolCall, ce.db, conversationID, messageId)
		if err != nil {
			return "", err
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
			return "", err
		}
	}
	conversationState.NextState(responseType)
	return botResponse, nil
}

func (ce *ConversationExecutor) prepareMessages(history []openai.ChatCompletionMessage, packages []external.Package, userInput string, whatsapp bool) []openai.ChatCompletionMessage {
	var systemTemplate string
	if whatsapp {
		systemTemplate = llm_service.SystemMessageTemplate(packages)
	} else {
		systemTemplate = llm_service.SystemMessageTemplateForWhatsapp(packages)
	}
	messages := append([]openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleSystem, Content: systemTemplate}}, history...)
	messages = append(messages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: userInput})
	return messages
}

func (ce *ConversationExecutor) getPackageListFromCache() ([]external.Package, error) {
	cacheKey := fmt.Sprintf(cache.CacheKeys.GetPackageList.Key)
	var packages []external.Package

	if err := cache.GetCache(cacheKey, &packages); err == nil {
		log.Println("Cache hit for package list")
		return packages, nil
	}

	log.Println("Cache miss for package list, fetching from external source")
	packages, err := external.GetPackageList()
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

func processFunctionResponse(toolCall openai.ToolCall, db *gorm.DB, conversationID uint, messageId uint) (*external.PackageDetails, error) {
	// Generalize the handling of function calls based on the function's name
	switch toolCall.Function.Name {
	case "get_package_details":
		return handleGetPackageDetails(toolCall, db, conversationID, messageId)
	// Add more cases as necessary for other function types
	default:
		log.Printf("Unhandled function call: %s", toolCall.Function.Name)
		return nil, nil
	}
}
