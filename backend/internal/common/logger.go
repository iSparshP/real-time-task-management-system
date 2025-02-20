package common

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

// InitLogger initializes the global logger
func InitLogger() error {
	config := zap.NewProductionConfig()

	// Customize the logging format
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)

	// Set log level based on environment
	if AppConfig.Environment == "development" {
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		config.Development = true
		config.Encoding = "console"
	} else {
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		config.Encoding = "json"
	}

	// Create the logger
	var err error
	Logger, err = config.Build(
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return err
	}

	// Replace the global logger
	zap.ReplaceGlobals(Logger)

	return nil
}

// Close properly syncs the logger before the application exits
func CloseLogger() {
	if Logger != nil {
		Logger.Sync()
	}
}
