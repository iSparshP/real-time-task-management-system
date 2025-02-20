package task

import (
	"time"
)

type MessageType string

const (
	MessageTypeTaskCreated  MessageType = "task_created"
	MessageTypeTaskUpdated  MessageType = "task_updated"
	MessageTypeTaskDeleted  MessageType = "task_deleted"
	MessageTypeTaskAssigned MessageType = "task_assigned"
)

type WebSocketMessage struct {
	Type      MessageType `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

func NewWebSocketMessage(msgType MessageType, payload interface{}) WebSocketMessage {
	return WebSocketMessage{
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now(),
	}
}
