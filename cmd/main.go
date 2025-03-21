package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"smart-chat/auth"
	"smart-chat/cache"
	"smart-chat/config"
	"smart-chat/external/notification"
	middleware "smart-chat/internal/middlewares"
	"smart-chat/internal/models"
	"smart-chat/internal/routes"
	"smart-chat/internal/services/conversation"
	convHistory "smart-chat/internal/services/conversation_history"
	notifications_job "smart-chat/internal/services/notifications_job"
	userService "smart-chat/internal/services/user"
	utils "smart-chat/internal/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cache.Initialize("localhost:11211")
	cfg := config.Load()

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=require TimeZone=Asia/Kolkata", cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Perform the conditional migration to add the "analysed" column if it doesn't exist
	err = applyConditionalMigrations(db)
	if err != nil {
		log.Fatalf("Failed to apply migration: %v", err)
	}

	// Perform other migrations (if necessary)
	err = db.AutoMigrate(&models.User{}, &models.Session{}, &models.Conversation{}, &models.MessagePair{}, &models.FunctionCall{}, &models.Button{}, &models.ConvAnalysis{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	router.Use(cors.New(config))

	v1 := router.Group("/v1")

	authService := auth.NewAuthService()
	authGroup := v1.Group("/auth")
	auth.RegisterAuthRoutes(authGroup, authService)

	chatGroup := v1.Group("/chat")
	chatGroup.Use(middleware.AuthMiddleware())
	routes.RegisterRoutes(chatGroup)

	v2 := router.Group("/v2")
	authServicev2 := auth.NewAuthV2Service(db)
	authGroupv2 := v2.Group("/auth")
	auth.RegisterV2AuthRoutes(authGroupv2, authServicev2)

	conversationService := conversation.NewConversationService(db)
	notifClient := notification.NewClient("http://127.0.0.1:8000")
	jobService := notifications_job.NewJobService(notifClient, db)

	chatGroupV2 := v2.Group("/chat")
	chatGroupV2.Use(middleware.AuthSessionMiddleware(db))
	routes.RegisterV2Routes(chatGroupV2, conversationService, jobService)

	conversationHistoryService := convHistory.NewConvHistoryService(db)
	us := userService.NewUserService(db)
	clientGroupV2 := v2.Group("/client")
	routes.ClientRoutes(clientGroupV2, conversationHistoryService, us)

	// Start cron jobs
	//cron_jobs.StartCronJobs(db)

	c := cron.New()
	if _, err := c.AddFunc("0 0 * * *", utils.PushConversationsToS3); err != nil {
		log.Fatalf("Failed to schedule cron job: %v", err)
	}
	c.Start()

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// applyConditionalMigrations finds all migration files in the migrations directory and applies them
func applyConditionalMigrations(db *gorm.DB) error {

	// Get the current file path and calculate the migrations directory path relative to the current file
	_, currentFilePath, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(currentFilePath)
	migrationsDir := filepath.Join(baseDir, "../migrations")

	// Open the migrations directory
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %v", err)
	}

	// Loop through all migration files in the directory
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".sql" {
			// For each migration SQL file, apply it
			migrationFilePath := filepath.Join(migrationsDir, file.Name())
			log.Printf("Applying migration: %s", file.Name())

			// Read the contents of the migration file
			migrationSQL, err := os.ReadFile(migrationFilePath) // Use os.ReadFile instead of ioutil.ReadFile
			if err != nil {
				log.Printf("Failed to read migration file %s: %v", file.Name(), err)
				continue
			}

			// Execute the migration using raw SQL
			err = db.Exec(string(migrationSQL)).Error
			if err != nil {
				log.Printf("Failed to apply migration %s: %v", file.Name(), err)
				return fmt.Errorf("failed to apply migration %s: %v", file.Name(), err)
			}
			log.Printf("Successfully applied migration: %s", file.Name())
		}
	}

	return nil
}
