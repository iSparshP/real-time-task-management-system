package notification

import (
	"github.com/iSparshP/real-time-task-management-system/internal/task"
)

type NotificationType string

const (
	NotificationTypeTaskCreated NotificationType = "task_created"
	NotificationTypeTaskUpdated NotificationType = "task_updated"
	NotificationTypeTaskDeleted NotificationType = "task_deleted"
	NotificationTypeTaskDue     NotificationType = "task_due"
)

type NotificationChannel string

const (
	ChannelSlack   NotificationChannel = "slack"
	ChannelDiscord NotificationChannel = "discord"
)

type NotificationConfig struct {
	SlackWebhookURL     string
	DiscordWebhookURL   string
	DefaultChannels     []NotificationChannel
	TaskUpdateThreshold int    // Minimum priority level for task update notifications
	DefaultUsername     string // Added for identifying the updater
}

type NotificationEvent struct {
	Type     NotificationType       `json:"type"`
	Task     task.Task              `json:"task"`
	Channels []NotificationChannel  `json:"channels,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type SlackBlock struct {
	Type     string              `json:"type"`
	Text     map[string]string   `json:"text,omitempty"`
	Elements []map[string]string `json:"elements,omitempty"`
}

type SlackPayload struct {
	Text   string       `json:"text"`
	Blocks []SlackBlock `json:"blocks"`
}

type DiscordEmbed struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Fields      []DiscordEmbedField `json:"fields"`
	Timestamp   string              `json:"timestamp"`
	Color       int                 `json:"color"`
}

type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type DiscordPayload struct {
	Content string         `json:"content"`
	Embeds  []DiscordEmbed `json:"embeds"`
}
