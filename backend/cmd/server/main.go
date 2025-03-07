package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/iSparshP/real-time-task-management-system/internal/ai"
	"github.com/iSparshP/real-time-task-management-system/internal/auth"
	"github.com/iSparshP/real-time-task-management-system/internal/common"
	"github.com/iSparshP/real-time-task-management-system/internal/database"
	"github.com/iSparshP/real-time-task-management-system/internal/notification"
	"github.com/iSparshP/real-time-task-management-system/internal/task"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Load application config
	if err := common.LoadConfig(); err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize logger
	if err := common.InitLogger(); err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	logger := common.Logger
	defer logger.Sync()

	// Initialize router with middleware
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(common.RequestLogger(logger))

	// Add after loading environment variables
	dbConfig := database.Config{
		Host:        os.Getenv("DB_HOST"),
		Port:        common.GetEnvInt("DB_PORT", 5432),
		User:        os.Getenv("DB_USER"),
		Password:    os.Getenv("DB_PASSWORD"),
		DBName:      os.Getenv("DB_NAME"),
		SSLMode:     os.Getenv("DB_SSLMODE"),
		ConnTimeout: 10 * time.Second,
		MaxRetries:  3,
	}

	db, err := database.NewGormDB(dbConfig)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer database.CloseDB(db)

	// Verify database connection
	if err := database.CheckConnection(db); err != nil {
		logger.Fatal("Database connection check failed", zap.Error(err))
	}

	// Run migrations
	if err := database.AutoMigrate(db); err != nil {
		logger.Fatal("Failed to run database migrations", zap.Error(err))
	}

	// Initialize services
	taskService := task.NewService(db, logger)
	taskHandler := task.NewHandler(taskService, logger)

	aiConfig := ai.AIProviderConfig{
		Provider:    os.Getenv("AI_PROVIDER"),
		APIKey:      os.Getenv("AI_API_KEY"),
		ModelName:   os.Getenv("AI_MODEL_NAME"),
		MaxTokens:   150,
		Temperature: 0.7,
	}
	aiService, err := ai.NewService(aiConfig, logger)
	if err != nil {
		logger.Fatal("Failed to initialize AI service", zap.Error(err))
	}
	aiHandler := ai.NewHandler(aiService, logger)

	notificationConfig := notification.NotificationConfig{
		SlackWebhookURL:   os.Getenv("SLACK_WEBHOOK_URL"),
		DiscordWebhookURL: os.Getenv("DISCORD_WEBHOOK_URL"),
		DefaultChannels: []notification.NotificationChannel{
			notification.ChannelSlack,
			notification.ChannelDiscord,
		},
	}
	notificationService, err := notification.NewService(notificationConfig, logger)
	if err != nil {
		logger.Fatal("Failed to initialize notification service", zap.Error(err))
	}
	defer notificationService.Close()
	notificationHandler := notification.NewHandler(notificationService, logger)

	authConfig := auth.Config{
		JWTSecret:              os.Getenv("JWT_SECRET"),
		TokenExpiration:        24 * time.Hour,
		RefreshTokenExpiration: 7 * 24 * time.Hour,
	}
	authService := auth.NewService(db, authConfig)
	authHandler := auth.NewHandler(authService, logger)

	// API routes - simplified structure
	api := router.Group("/api")
	{
		// Unprotected routes
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/refresh", authHandler.RefreshToken)

		// Protected routes
		api.Use(auth.AuthMiddleware(authService))
		{
			// Task routes
			api.GET("/tasks/ws", taskHandler.WebSocket)
			api.POST("/tasks", taskHandler.CreateTask)
			api.GET("/tasks", taskHandler.ListTasks)
			api.GET("/tasks/:id", taskHandler.GetTask)
			api.PUT("/tasks/:id", taskHandler.UpdateTask)
			api.DELETE("/tasks/:id", taskHandler.DeleteTask)
			api.POST("/tasks/:id/assign", taskHandler.AssignTask)

			// AI routes
			api.POST("/ai/suggest", aiHandler.GetSuggestions)

			// Notification routes
			api.POST("/notifications/events", notificationHandler.HandleTaskEvent)
		}
	}

	// Server configuration
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", os.Getenv("PORT")),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exiting")
}
