package utils

import (
	"log"
	"smart-chat/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupTestDB initializes a new SQLite database for testing.
// It returns a GORM DB connection and a teardown function.
func SetupTestDB() (*gorm.DB, func()) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to the database: %v", err)
	}

	// Migrate the schema
	if err := db.AutoMigrate(&models.User{}, &models.Session{}, &models.Conversation{}, &models.MessagePair{}, &models.FunctionCall{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// Teardown function to clean up after tests
	teardown := func() {
		sqlDB, err := db.DB()
		if err != nil {
			log.Fatalf("failed to close the database: %v", err)
		}
		sqlDB.Close()
	}

	return db, teardown
}
