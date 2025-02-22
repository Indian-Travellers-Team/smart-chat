package llm_service

import (
	"context"
	"log"
	"smart-chat/internal/models"

	openai "github.com/sashabaranov/go-openai"
)

func ConvAnalysisTemplate() string {
	// Return the static HTML template directly with instructions
	return `
		<html>
		<head>
			<title>Conversation Analysis</title>
			<style>
				body { font-family: Arial, sans-serif; }
				.container { margin: 20px; }
				.question { font-weight: bold; }
				.answer { color: #333; }
				.summary { font-style: italic; color: #555; }
			</style>
		</head>
		<body>
			<div class="container">
				<h2>Conversation Analysis</h2>
				<p class="question">1. What package is the user interested in?</p>
				<p class="answer"><b>{{user_package}}</b></p>

				<p class="question">2. How many people is the user booking the trip for?</p>
				<p class="answer"><b>{{num_people}}</b></p>

				<p class="question">3. What is the date of the booking the user is interested in?</p>
				<p class="answer"><b>{{booking_date}}</b></p>

				<p class="question">4. Did the user complete the conversation?</p>
				<p class="answer"><b>{{conversation_complete}}</b></p>

				<p class="question">5. Did the user look satisfied by the end of the conversation?</p>
				<p class="answer"><b>{{user_satisfaction}}</b></p>

				<p class="question">6. Short summary of the conversation:</p>
				<p class="summary">
					<b>{{summary}}</b>
				</p>
			</div>
		</body>
		</html>
	`
}

func GetConversationSummary(conversation models.Conversation) (string, error) {
	// Get the HTML template for conversation analysis
	systemTemplate := ConvAnalysisTemplate() // Use the static string template directly

	// Instructions for OpenAI to generate the conversation analysis in the specified HTML format
	instructions := `
		You are an assistant tasked with analyzing conversations between a user and a travel booking assistant.
		Your goal is to analyze the conversation and provide the information in the following HTML format:

		1. What package is the user interested in? (e.g., "Chopta Tungnath")
		2. How many people is the user booking the trip for? (e.g., "3 people")
		3. What is the date of the booking the user is interested in? (e.g., "15th December 2024")
		4. Did the user complete the conversation? (e.g., "Yes, the user seemed to end it by saying goodbye")
		5. Did the user look satisfied by the end of the conversation? (e.g., "Yes, the user seemed satisfied because they were asking more questions")
		6. Provide a short summary of the conversation.

		Ensure that your response includes the correct information in the HTML format provided above. Do not modify the HTML structure. Only replace the placeholder text (e.g., {{user_package}}, {{num_people}}, etc.) with the actual conversation analysis.
	`

	var messages []openai.ChatCompletionMessage
	// Add instructions and the system template
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: instructions + systemTemplate,
	})

	// Append the conversation message pairs
	for _, pair := range conversation.MessagePairs {
		if pair.Visible {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: pair.User,
			}, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: pair.Bot,
			})
		}
	}

	// Make the OpenAI request to generate a conversation summary
	ctx := context.Background()
	req := openai.ChatCompletionRequest{
		Model:    openai.GPT4oMini,
		Messages: messages,
	}
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Printf("Error creating chat completion: %v", err)
		return "", err
	}
	log.Printf("token usage: %v", resp.Usage.TotalTokens)

	// Return the generated HTML content from OpenAI
	return resp.Choices[0].Message.Content, nil
}
