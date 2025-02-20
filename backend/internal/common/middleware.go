package common

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RequestIDKey is the key used to store request ID in context
const RequestIDKey = "RequestID"

// CORS middleware
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RequestID middleware adds a unique ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set(RequestIDKey, requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func RequestLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		logger.Info("Request",
			zap.String("path", path),
			zap.String("method", c.Request.Method),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
		)
	}
}
