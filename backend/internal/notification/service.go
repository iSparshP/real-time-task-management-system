package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Service struct {
	config NotificationConfig
	logger *zap.Logger
	client *http.Client
	wg     sync.WaitGroup
}

func NewService(config NotificationConfig, logger *zap.Logger) (*Service, error) {
	return &Service{
		config: config,
		logger: logger,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

func (s *Service) SendNotification(event NotificationEvent) {
	channels := event.Channels
	if len(channels) == 0 {
		channels = s.config.DefaultChannels
	}

	for _, channel := range channels {
		s.wg.Add(1)
		go func(ch NotificationChannel) {
			defer s.wg.Done()

			var err error
			switch ch {
			case ChannelSlack:
				err = s.sendSlackNotification(event)
			case ChannelDiscord:
				err = s.sendDiscordNotification(event)
			}

			if err != nil {
				s.logger.Error("Failed to send notification",
					zap.String("channel", string(ch)),
					zap.Error(err),
				)
			}
		}(channel)
	}
}

func (s *Service) sendSlackNotification(event NotificationEvent) error {
	if s.config.SlackWebhookURL == "" {
		return fmt.Errorf("slack webhook URL not configured")
	}

	// Create Slack-specific payload
	blocks := []map[string]interface{}{
		{
			"type": "section",
			"text": map[string]interface{}{
				"type": "mrkdwn",
				"text": fmt.Sprintf("*Task Update*\n*Task:* %s\n*Updated by:* %s\n*Status:* %s",
					event.Task.Title,
					event.Task.CreatedBy,
					event.Task.Status),
			},
		},
		{
			"type": "context",
			"elements": []map[string]interface{}{
				{
					"type": "mrkdwn",
					"text": fmt.Sprintf("Timestamp: %s", time.Now().Format(time.RFC3339)),
				},
			},
		},
	}

	payload := map[string]interface{}{
		"text":   fmt.Sprintf("Task Update: Task '%s' has been updated.", event.Task.Title),
		"blocks": blocks,
	}

	return s.sendWebhookRequest(s.config.SlackWebhookURL, payload)
}

func (s *Service) sendDiscordNotification(event NotificationEvent) error {
	if s.config.DiscordWebhookURL == "" {
		return fmt.Errorf("discord webhook URL not configured")
	}

	// Create Discord-specific payload
	embed := map[string]interface{}{
		"title":       fmt.Sprintf("Task Update: %s", event.Task.Title),
		"description": "The task has been updated.",
		"fields": []map[string]interface{}{
			{
				"name":   "Updated by",
				"value":  event.Task.CreatedBy,
				"inline": true,
			},
			{
				"name":   "Status",
				"value":  string(event.Task.Status),
				"inline": true,
			},
		},
		"timestamp": time.Now().Format(time.RFC3339),
		"color":     s.getDiscordColorForEvent(event),
	}

	payload := map[string]interface{}{
		"content": "Task Update Notification",
		"embeds":  []interface{}{embed},
	}

	return s.sendWebhookRequest(s.config.DiscordWebhookURL, payload)
}

func (s *Service) sendWebhookRequest(webhookURL string, payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook request failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (s *Service) getNotificationTitle(event NotificationEvent) string {
	switch event.Type {
	case NotificationTypeTaskCreated:
		return "🆕 New Task Created"
	case NotificationTypeTaskUpdated:
		return "📝 Task Updated"
	case NotificationTypeTaskDeleted:
		return "🗑️ Task Deleted"
	case NotificationTypeTaskDue:
		return "⏰ Task Due Soon"
	default:
		return "Task Notification"
	}
}

func (s *Service) getColorForEvent(event NotificationEvent) string {
	switch event.Type {
	case NotificationTypeTaskCreated:
		return "#36a64f" // green
	case NotificationTypeTaskUpdated:
		return "#2196f3" // blue
	case NotificationTypeTaskDeleted:
		return "#f44336" // red
	case NotificationTypeTaskDue:
		return "#ff9800" // orange
	default:
		return "#9e9e9e" // grey
	}
}

func (s *Service) getDiscordColorForEvent(event NotificationEvent) int {
	switch event.Type {
	case NotificationTypeTaskCreated:
		return 3066993 // Green
	case NotificationTypeTaskUpdated:
		return 5814783 // Blue
	case NotificationTypeTaskDeleted:
		return 15158332 // Red
	case NotificationTypeTaskDue:
		return 16776960 // Yellow
	default:
		return 10197915 // Gray
	}
}

func (s *Service) Close() {
	s.wg.Wait()
}
