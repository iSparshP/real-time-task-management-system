package common

// EventType represents different types of system events
type EventType string

const (
	EventTaskCreated EventType = "task_created"
	EventTaskUpdated EventType = "task_updated"
	EventTaskDeleted EventType = "task_deleted"
	EventTaskDue     EventType = "task_due"
	EventError       EventType = "error"
)

// Event represents a system event with payload
type Event struct {
	Type    EventType   `json:"type"`
	Payload interface{} `json:"payload"`
}
