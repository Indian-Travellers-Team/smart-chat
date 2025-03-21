package handlers

import (
	"encoding/json"
	"log"
	external "smart-chat/external/indian_travellers"
	"smart-chat/internal/llm_service"
	"smart-chat/internal/store"

	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"
)

func SendMessageHandler(c *gin.Context) {
	accessToken := c.GetHeader("Authorization")

	conversation, _ := store.GetConversation(accessToken)

	// Check if the conversation already has 10 message pairs
	if len(conversation.Messages) >= 10 {
		c.JSON(500, gin.H{"error": "Exceeded 10 of conversations"})
		return
	}

	var jsonData struct {
		Message string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&jsonData); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})

	}
	packages, nil := external.GetPackageList()
	var messages []openai.ChatCompletionMessage
	systemTemplate := llm_service.SystemMessageTemplate(packages)
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: systemTemplate,
	})

	for _, pair := range conversation.Messages {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: pair.UserMessage,
		}, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: pair.BotMessage,
		})
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: jsonData.Message,
	})

	typ, response, usedTokens, err := llm_service.GetOpenAIResponse(messages)
	if err != nil {
		log.Printf("Error getting response from OpenAI: %v", err)
		c.JSON(500, gin.H{"error": "error processing message"})
		return
	}
	if typ == "function" {
		functionDetails := response.(openai.ToolCall)
		type Package struct {
			PackageId int `json:"package_id"`
		}
		var p Package
		err := json.Unmarshal([]byte(functionDetails.Function.Arguments), &p)
		if err != nil {
			log.Printf("Error getting response from OpenAI: %v", err)
			c.JSON(500, gin.H{"error": "error processing message"})
			return
		}
		packageDetail, _ := external.GetPackageDetails(p.PackageId)
		packageDetailString, _ := json.Marshal(packageDetail)
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleFunction,
			Content: string(packageDetailString),
			Name:    functionDetails.Function.Name,
		})

		_, response, usedTokens, llm_err := llm_service.GetOpenAIResponse(messages)

		if llm_err != nil {
			c.JSON(500, gin.H{"error": "error with conversation"})
			return
		}
		messagePair := store.MessagePair{
			UserMessage: jsonData.Message,
			BotMessage:  response.(string),
		}
		if err := store.AppendToConversation(accessToken, messagePair, usedTokens); err != nil {
			c.JSON(500, gin.H{"error": "error updating conversation"})
			return
		}

		c.JSON(200, gin.H{
			"status":   "message sent and processed",
			"response": response,
		})
	} else {
		messagePair := store.MessagePair{
			UserMessage: jsonData.Message,
			BotMessage:  response.(string),
		}
		if err := store.AppendToConversation(accessToken, messagePair, usedTokens); err != nil {
			c.JSON(500, gin.H{"error": "error updating conversation"})
			return
		}

		c.JSON(200, gin.H{
			"status":   "message sent and processed",
			"response": response,
		})
	}
}
