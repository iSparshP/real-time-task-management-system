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
	SlackToken          string
	SlackChannel        string
	DiscordToken        string
	DiscordChannelID    string
	DefaultChannels     []NotificationChannel
	TaskUpdateThreshold int // Minimum priority level for task update notifications
}

type NotificationEvent struct {
	Type     NotificationType       `json:"type"`
	Task     task.Task              `json:"task"`
	Channels []NotificationChannel  `json:"channels,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
