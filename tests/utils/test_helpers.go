package utils

import (
	"log"
	"smart-chat/internal/models"
	"time"

	"gorm.io/gorm"
)

func SetupTestEntities(db *gorm.DB) (models.User, models.Session, models.Conversation, models.MessagePair) {
	user := models.User{
		Name:           "Test User",
		Mobile:         "1234567890",
		OTP:            "1234",
		AccessToken:    "access_token_sample",
		AccessExpireAt: time.Now().Add(24 * time.Hour),
	}
	if err := db.Create(&user).Error; err != nil {
		log.Fatalf("Creating user failed: %v", err)
	}

	authToken := "valid_token"
	session := models.Session{
		UserID:    user.ID,
		AuthToken: authToken,
		ExpireAt:  time.Now().Add(24 * time.Hour),
	}
	if err := db.Create(&session).Error; err != nil {
		log.Fatalf("Creating session failed: %v", err)
	}

	conversation := models.Conversation{
		SessionID: session.ID,
	}
	if err := db.Create(&conversation).Error; err != nil {
		log.Fatalf("Creating conversation failed: %v", err)
	}

	messagePair := models.MessagePair{
		ConversationID: conversation.ID,
		User:           "Hello",
		Bot:            "Hi there!",
		Visible:        true,
	}
	if err := db.Create(&messagePair).Error; err != nil {
		log.Fatalf("Creating message pair failed: %v", err)
	}

	return user, session, conversation, messagePair
}
