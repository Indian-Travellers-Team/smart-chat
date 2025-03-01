// llm_service/executor.go
package llm_service

import (
	"context"
	"encoding/json"
	"log"
	"smart-chat/config"
	"smart-chat/internal/models"
	"strings"

	"github.com/jaytaylor/html2text"
	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

var client *openai.Client

func init() {
	cfg := config.Load()
	openAIToken := cfg.OpenAIKey
	if openAIToken == "" {
		log.Fatal("OPENAI_API_KEY is not set in environment variables")
	}
	client = openai.NewClient(openAIToken)
}

func GetOpenAIResponse(messages []openai.ChatCompletionMessage) (string, interface{}, int, error) {
	ctx := context.Background()
	var tools = []openai.Tool{
		{Type: "function", Function: GetPackageDetailsSchema},
	}
	req := openai.ChatCompletionRequest{
		Model:    openai.GPT4o,
		Messages: messages,
		Tools:    tools,
	}
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Printf("Error creating chat completion: %v", err)
		return "msg", "Service currently unavailable", 0, err
	}

	totalTokens := resp.Usage.TotalTokens

	msg := resp.Choices[0].Message
	if len(msg.ToolCalls) != 0 {
		functionName := msg.ToolCalls[0].Function.Name
		log.Printf("function call %v", functionName)
		return "function", msg.ToolCalls[0], totalTokens, nil
	}

	if len(resp.Choices) == 0 || len(msg.Content) == 0 {
		return "msg", "Service currently unavailable", totalTokens, nil
	}

	return "msg", resp.Choices[0].Message.Content, totalTokens, nil
}

func GetOpenAIResponsev2(messages []openai.ChatCompletionMessage) (models.MessageType, interface{}, uint, error) {
	ctx := context.Background()
	var tools = []openai.Tool{
		{Type: "function", Function: GetPackageDetailsSchema},
	}

	// Define the schema for the response
	type ResponseSchema struct {
		Content string   `json:"content"`
		Hints   []string `json:"hints"`
	}

	schema, _ := jsonschema.GenerateSchemaForType(ResponseSchema{})

	req := openai.ChatCompletionRequest{
		Model:    openai.GPT4o,
		Messages: messages,
		Tools:    tools,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:   "response_with_hints",
				Schema: schema,
				Strict: true,
			},
		},
	}
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Printf("Error creating chat completion: %v", err)
		return models.MessageTypeUserSent, "Service currently unavailable", 0, err
	}
	log.Printf("token usage: %v", resp.Usage.TotalTokens)
	totalTokens := uint(resp.Usage.TotalTokens)

	msg := resp.Choices[0].Message
	if len(msg.ToolCalls) != 0 {
		functionName := msg.ToolCalls[0].Function.Name
		log.Printf("function call %v", functionName)
		return models.MessageTypeFunctionCall, msg.ToolCalls[0], totalTokens, nil
	}

	if len(resp.Choices) == 0 || len(msg.Content) == 0 {
		return models.MessageTypeUserSent, "Service currently unavailable", totalTokens, nil
	}

	return models.MessageTypeUserSent, resp.Choices[0].Message.Content, totalTokens, nil
}

func GetOpenAIResponsev2Whatsapp(messages []openai.ChatCompletionMessage) (models.MessageType, interface{}, uint, error) {
	ctx := context.Background()
	var tools = []openai.Tool{
		{Type: "function", Function: GetPackageDetailsSchema},
		{Type: "function", Function: CreateUserInitialQuerySchema},
		{Type: "function", Function: CreateUserFinalBookingSchema},
		{Type: "function", Function: FetchUpcomingTripsSchema},
	}

	// Define the schema for the response
	type ResponseSchema struct {
		Content string `json:"content"`
	}

	schema, _ := jsonschema.GenerateSchemaForType(ResponseSchema{})

	req := openai.ChatCompletionRequest{
		Model:    openai.GPT4o,
		Messages: messages,
		Tools:    tools,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:   "response_without_hints",
				Schema: schema,
				Strict: true,
			},
		},
	}
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Printf("Error creating chat completion: %v", err)
		return models.MessageTypeUserSent, "Service currently unavailable", 0, err
	}
	log.Printf("token usage: %v", resp.Usage.TotalTokens)
	totalTokens := uint(resp.Usage.TotalTokens)

	msg := resp.Choices[0].Message
	if len(msg.ToolCalls) != 0 {
		functionName := msg.ToolCalls[0].Function.Name
		log.Printf("function call %v", functionName)
		return models.MessageTypeFunctionCall, msg.ToolCalls[0], totalTokens, nil
	}

	if len(resp.Choices) == 0 || len(msg.Content) == 0 {
		return models.MessageTypeUserSent, "Service currently unavailable", totalTokens, nil
	}
	var result map[string]interface{}

	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result)
	if err != nil {
		log.Fatal(err)
	}

	markdown := result["content"].(string)
	text, _ := html2text.FromString(string(markdown))

	text = strings.ReplaceAll(text, "**", "*")
	text = strings.ReplaceAll(text, "### ", "*")

	result["content"] = text

	var stringContent []byte

	stringContent, err = json.Marshal(result)
	if err != nil {
		log.Fatal(err)
	}
	return models.MessageTypeUserSent, string(stringContent), totalTokens, nil
}
