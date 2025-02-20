package ai

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	service *Service
	logger  *zap.Logger
}

func NewHandler(service *Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) GetSuggestions(c *gin.Context) {
	var req SuggestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid suggestion request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	if err := h.validateRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": err.Error(),
		})
		return
	}

	resp, err := h.service.GetSuggestions(req)
	if err != nil {
		switch {
		case errors.Is(err, ErrRateLimitExceeded):
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": "60s",
			})
		case errors.Is(err, ErrRateLimit):
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "AI provider rate limit exceeded",
				"retry_after": "30s",
			})
		case errors.Is(err, ErrQuota):
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "AI provider quota exceeded",
				"message": "Please contact support to increase your quota",
			})
		case errors.Is(err, ErrAIProviderUnavailable):
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":       "AI service temporarily unavailable",
				"retry_after": "30s",
			})
		case errors.Is(err, ErrInvalidResponse):
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to process AI response",
			})
		default:
			h.logger.Error("Failed to get AI suggestions",
				zap.Error(err),
				zap.String("task_id", req.Task.ID),
				zap.String("suggest_for", req.SuggestFor),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) validateRequest(req SuggestionRequest) error {
	if req.Task.Title == "" {
		return errors.New("task title is required")
	}

	if req.SuggestFor == "" {
		return errors.New("suggest_for field is required")
	}

	validSuggestionTypes := map[string]bool{
		"priority": true,
		"deadline": true,
		"approach": true,
	}

	if !validSuggestionTypes[req.SuggestFor] {
		return errors.New("invalid suggestion type")
	}

	if len(req.Task.Title) < 3 || len(req.Task.Title) > 200 {
		return errors.New("title length must be between 3 and 200 characters")
	}

	if len(req.Task.Description) > 1000 {
		return errors.New("description length must not exceed 1000 characters")
	}

	return nil
}
