package database

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/iSparshP/real-time-task-management-system/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Host        string
	Port        int
	User        string
	Password    string
	DBName      string
	SSLMode     string
	ConnTimeout time.Duration // Add connection timeout
	MaxRetries  int
}

func CheckConnection(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

func NewGormDB(config Config) (*gorm.DB, error) {
	if config.ConnTimeout == 0 {
		config.ConnTimeout = 10 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.SSLMode == "" {
		config.SSLMode = "require" // Default to require SSL for cloud databases
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s connect_timeout=%d",
		config.Host,
		config.User,
		config.Password,
		config.DBName,
		config.Port,
		config.SSLMode,
		int(config.ConnTimeout.Seconds()),
	)

	gormConfig := &gorm.Config{
		Logger: logger.New(
			log.New(log.Writer(), "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Info,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		),
		PrepareStmt: true, // Enable prepared statement cache
	}

	// Enhanced retry logic with exponential backoff
	var db *gorm.DB
	var err error
	for i := 0; i < config.MaxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
		if err == nil {
			if err := CheckConnection(db); err == nil {
				break
			}
		}
		if i < config.MaxRetries-1 {
			backoffDuration := time.Second * time.Duration(math.Pow(2, float64(i)))
			time.Sleep(backoffDuration)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d retries: %w", config.MaxRetries, err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Monitor connection health
	go monitorDBConnection(db)

	return db, nil
}

func monitorDBConnection(db *gorm.DB) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := CheckConnection(db); err != nil {
			log.Printf("Database connection check failed: %v", err)
		}
	}
}

func CloseDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	return nil
}

// AutoMigrate runs database migrations for all models
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Task{},
	)
}
