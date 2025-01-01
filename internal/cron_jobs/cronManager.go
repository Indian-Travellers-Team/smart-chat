package cron_jobs

import (
	"log"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

func StartCronJobs(db *gorm.DB) {
	// Create a new cron scheduler
	c := cron.New()

	// Add the conversation analysis job that runs every minute
	_, err := c.AddFunc("*/1 * * * *", func() {
		log.Println("Starting conversation analysis job...")
		GenerateConversationAnalysis(db)
	})

	if err != nil {
		log.Fatalf("Failed to schedule conv analyses job: %v", err)
	}

	recipients := []string{
		"arundeepak92@gmail.com",
		"deepak.deep40@yahoo.com",
		"bhartendumehta206@gmail.com",
	}

	_, err = c.AddFunc("*/1 * * * *", func() {
		log.Println("Starting email notification job...")
		EmailNotificationJob(db, recipients)
	})

	if err != nil {
		log.Fatalf("Failed to schedule email notification job with error: %v", err)
	}

	// Start the cron scheduler
	c.Start()
}
