package cron_jobs

import (
	"fmt"
	"log"
	"smart-chat/config"
	"smart-chat/internal/models"
	"smart-chat/internal/utils"
	"strings"

	"gorm.io/gorm"
)

var cfg = config.Load()

// EmailNotificationJob sends an email with summaries of unanalyzed conversations
func EmailNotificationJob(db *gorm.DB, recipients []string) error {
	// Get all unprocessed ConvAnalysis entries where email_sent = false
	var analyses []models.ConvAnalysis
	if err := db.Where("email_sent = ?", false).Find(&analyses).Error; err != nil {
		return fmt.Errorf("error fetching unprocessed conv_analyses entries: %v", err)
	}

	// If there are any unprocessed entries, send the email
	if len(analyses) > 0 {
		log.Printf("Found %d unprocessed ConvAnalysis entries. Sending email...", len(analyses))

		// Generate the email body
		emailBody := generateConvAnalysisEmailBody(db, analyses)

		// Send the email using the utils package
		recipientsStr := strings.Join(recipients, ",") // Convert the list of recipients to a comma-separated string
		if err := utils.SendEmail(cfg.Email, cfg.EmailPassword, recipientsStr, "Daily Conversation Analysis Summary", emailBody); err != nil {
			return fmt.Errorf("failed to send email: %v", err)
		}

		// Update the email_sent field to true for the processed entries
		for _, analysis := range analyses {
			analysis.EmailSent = true
			if err := db.Save(&analysis).Error; err != nil {
				log.Printf("Failed to update email_sent for conv_analysis ID %d: %v", analysis.ID, err)
			}
		}

		log.Println("Email sent successfully and conv_analyses updated.")
	} else {
		log.Println("No unprocessed conv_analyses found.")
	}

	return nil
}

// generateConvAnalysisEmailBody generates the email body for the unprocessed analyses
func generateConvAnalysisEmailBody(db *gorm.DB, analyses []models.ConvAnalysis) string {
	var emailBody string

	for _, analysis := range analyses {
		// Fetch user information related to the conversation's session
		var conversation models.Conversation
		if err := db.Preload("Session.User").First(&conversation, analysis.ConversationID).Error; err != nil {
			log.Printf("Error fetching conversation with ID %d: %v", analysis.ConversationID, err)
			continue
		}

		user := conversation.Session.User
		// Add user details to the email body
		emailBody += fmt.Sprintf("Conversation ID: %d\n", analysis.ConversationID)
		emailBody += fmt.Sprintf("User Name: %s\n", user.Name)
		emailBody += fmt.Sprintf("User Contact: %s\n", user.Mobile)
		emailBody += fmt.Sprintf("Summary: %s\n\n", analysis.Summary)
	}

	return emailBody
}
