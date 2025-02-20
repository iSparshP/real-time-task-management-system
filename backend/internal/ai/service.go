package ai

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"google.golang.org/api/option"
)

var (
	ErrAIProviderUnavailable = errors.New("AI provider unavailable")
	ErrInvalidResponse       = errors.New("invalid response from AI provider")
	ErrRateLimitExceeded     = errors.New("rate limit exceeded")
	ErrRateLimit             = errors.New("AI provider rate limit exceeded")
	ErrQuota                 = errors.New("AI provider quota exceeded")
)

type Service struct {
	client      *genai.Client
	model       *genai.GenerativeModel
	config      AIProviderConfig
	logger      *zap.Logger
	cache       *cache.Cache
	rateLimiter *rate.Limiter
	maxRetries  int
	retryDelay  time.Duration
}

func NewService(config AIProviderConfig, logger *zap.Logger) (*Service, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(config.APIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	model := client.GenerativeModel(config.ModelName)
	model.SetTemperature(config.Temperature)

	return &Service{
		client:      client,
		model:       model,
		config:      config,
		logger:      logger,
		cache:       cache.New(5*time.Minute, 10*time.Minute),
		rateLimiter: rate.NewLimiter(rate.Every(time.Second), 10),
		maxRetries:  3,
		retryDelay:  1 * time.Second,
	}, nil
}

func (s *Service) GetSuggestions(req SuggestionRequest) (*SuggestionResponse, error) {
	if !s.rateLimiter.Allow() {
		return nil, ErrRateLimitExceeded
	}

	// Check cache
	if cached, found := s.cache.Get(s.getCacheKey(req)); found {
		return cached.(*SuggestionResponse), nil
	}

	var lastErr error
	for attempt := 0; attempt <= s.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(s.getRetryDelay(attempt))
		}

		resp, err := s.makeAIRequest(req)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		if !s.shouldRetry(err) {
			break
		}

		s.logger.Warn("AI request failed, retrying",
			zap.Error(err),
			zap.Int("attempt", attempt+1),
			zap.Int("max_retries", s.maxRetries),
		)
	}

	return nil, fmt.Errorf("AI completion error after %d retries: %w", s.maxRetries, lastErr)
}

func (s *Service) makeAIRequest(req SuggestionRequest) (*SuggestionResponse, error) {
	ctx := context.Background()
	prompt := s.buildPrompt(req)

	resp, err := s.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		if strings.Contains(err.Error(), "quota") {
			return nil, ErrQuota
		}
		if strings.Contains(err.Error(), "rate") {
			return nil, ErrRateLimit
		}
		return nil, err
	}

	if len(resp.Candidates) == 0 {
		return nil, ErrInvalidResponse
	}

	// Get text from the response
	suggestion := ""
	if textPart, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		suggestion = string(textPart)
	} else {
		return nil, ErrInvalidResponse
	}

	confidence := 1.0
	if resp.Candidates[0].FinishReason == genai.FinishReasonMaxTokens {
		confidence = 0.0
	}

	response := &SuggestionResponse{
		Suggestions: []Suggestion{
			{
				Type:       "primary",
				Suggestion: suggestion,
				Confidence: math.Round(confidence*100) / 100,
			},
		},
	}

	// Cache the response
	s.cache.Set(s.getCacheKey(req), response, cache.DefaultExpiration)

	return response, nil
}

func (s *Service) shouldRetry(err error) bool {
	return err == ErrRateLimit || strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "connection refused")
}

func (s *Service) getRetryDelay(attempt int) time.Duration {
	return s.retryDelay * time.Duration(math.Pow(2, float64(attempt-1)))
}

func (s *Service) buildPrompt(req SuggestionRequest) string {
	var prompt string
	switch req.SuggestFor {
	case "priority":
		prompt = fmt.Sprintf(
			"Given the following task details:\nTitle: %s\nDescription: %s\nDue Date: %s\n"+
				"Please suggest an appropriate priority level (low/medium/high) and provide reasoning.\n"+
				"Consider task complexity, due date, and impact.",
			req.Task.Title, req.Task.Description, req.Task.DueDate.Format("2006-01-02"),
		)
	case "deadline":
		prompt = fmt.Sprintf(
			"For the following task:\nTitle: %s\nDescription: %s\nPriority: %s\n"+
				"Suggest an appropriate deadline considering the task complexity and priority.\n"+
				"Provide reasoning for the suggested deadline.",
			req.Task.Title, req.Task.Description, req.Task.Priority,
		)
	case "approach":
		prompt = fmt.Sprintf(
			"For the task:\nTitle: %s\nDescription: %s\n"+
				"Suggest the best approach to complete this task efficiently.\n"+
				"Consider breaking it down into smaller steps if appropriate.",
			req.Task.Title, req.Task.Description,
		)
	}

	if req.UserContext != "" {
		prompt += fmt.Sprintf("\nAdditional context: %s", req.UserContext)
	}

	return prompt
}

func (s *Service) getCacheKey(req SuggestionRequest) string {
	return fmt.Sprintf("%s:%s:%s", req.Task.ID, req.SuggestFor, req.UserContext)
}
