package common

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Database settings
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string

	// Redis settings
	RedisHost     string
	RedisPort     int
	RedisPassword string
	RedisDB       int

	// Server settings
	ServerPort  int
	Environment string

	// Task settings
	TaskDefaultStatus string
	TaskPageSize      int
	TaskMaxDescLength int
}

var AppConfig Config

// LoadConfig loads all environment variables into the Config struct
func LoadConfig() error {
	if err := godotenv.Load(); err != nil {
		// Only log warning as .env file is optional
		fmt.Println("Warning: .env file not found")
	}

	// Database configuration
	AppConfig.DBHost = getEnvString("DB_HOST", "localhost")
	AppConfig.DBPort = GetEnvInt("DB_PORT", 5432)
	AppConfig.DBUser = getEnvString("DB_USER", "postgres")
	AppConfig.DBPassword = getEnvString("DB_PASSWORD", "")
	AppConfig.DBName = getEnvString("DB_NAME", "app_db")

	// Redis configuration
	AppConfig.RedisHost = getEnvString("REDIS_HOST", "localhost")
	AppConfig.RedisPort = GetEnvInt("REDIS_PORT", 6379)
	AppConfig.RedisPassword = getEnvString("REDIS_PASSWORD", "")
	AppConfig.RedisDB = GetEnvInt("REDIS_DB", 0)

	// Server configuration
	AppConfig.ServerPort = GetEnvInt("SERVER_PORT", 8080)
	AppConfig.Environment = getEnvString("ENVIRONMENT", "development")

	// Task configuration
	AppConfig.TaskDefaultStatus = getEnvString("TASK_DEFAULT_STATUS", "pending")
	AppConfig.TaskPageSize = GetEnvInt("TASK_PAGE_SIZE", 10)
	AppConfig.TaskMaxDescLength = GetEnvInt("TASK_MAX_DESCRIPTION_LENGTH", 1000) // Ensure this is set

	if AppConfig.TaskMaxDescLength <= 0 {
		AppConfig.TaskMaxDescLength = 1000 // Fallback default if environment variable is invalid
	}

	return nil
}

// Helper functions to get environment variables with default values
func getEnvString(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func GetEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
