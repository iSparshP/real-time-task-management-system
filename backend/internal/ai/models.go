package ai

import (
	"time"

	"github.com/iSparshP/real-time-task-management-system/internal/task"
)

type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Priority    string    `json:"priority"`
	DueDate     time.Time `json:"due_date"`
}

type SuggestionRequest struct {
	Task        task.Task `json:"task"`
	SuggestFor  string    `json:"suggest_for" binding:"required,oneof=priority deadline approach"`
	UserContext string    `json:"user_context,omitempty"`
}

type Suggestion struct {
	Type       string  `json:"type"`
	Suggestion string  `json:"suggestion"`
	Reasoning  string  `json:"reasoning"`
	Confidence float64 `json:"confidence"`
}

type SuggestionResponse struct {
	Suggestions []Suggestion `json:"suggestions"`
}

type AIProviderConfig struct {
	Provider    string  `json:"provider"`
	APIKey      string  `json:"api_key"`
	ModelName   string  `json:"model_name"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float32 `json:"temperature"`
}
