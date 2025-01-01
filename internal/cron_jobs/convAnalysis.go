package cron_jobs

import (
	"log"
	"smart-chat/internal/llm_service"
	"smart-chat/internal/models"

	"gorm.io/gorm"
)

func GenerateConversationAnalysis(db *gorm.DB) {
	// Fetch conversations that are not yet analyzed
	var conversations []models.Conversation

	err := db.Where("analysed = ?", false).Preload("MessagePairs").Find(&conversations).Error
	if err != nil {
		log.Printf("Error fetching unanalyzed conversations: %v", err)
		return
	}

	// Loop over each conversation and generate an analysis
	for _, conversation := range conversations {
		// Generate summary for each conversation using llm_service
		summary, err := llm_service.GetConversationSummary(conversation)
		if err != nil {
			log.Printf("Error generating summary for conversation %d: %v", conversation.ID, err)
			continue
		}

		// Store the analysis in the ConvAnalysis table
		convAnalysis := models.ConvAnalysis{
			ConversationID: conversation.ID,
			Summary:        summary,
			EmailSent:      false, // Assuming the email will be sent later
		}
		if err := db.Create(&convAnalysis).Error; err != nil {
			log.Printf("Error storing analysis for conversation %d: %v", conversation.ID, err)
		}

		// Mark the conversation as analyzed
		conversation.Analysed = true
		if err := db.Save(&conversation).Error; err != nil {
			log.Printf("Error updating conversation status for %d: %v", conversation.ID, err)
		}

		log.Printf("Conversation %d analyzed and stored", conversation.ID)
	}
}
