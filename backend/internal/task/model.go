package task

import (
	"time"

	"github.com/iSparshP/real-time-task-management-system/internal/models"
)

// Use models types directly
type Task = models.Task
type TaskStatus = models.TaskStatus
type TaskPriority = models.TaskPriority

// Request/response types
type CreateTaskRequest struct {
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	Priority    string    `json:"priority" binding:"required"`
	AssignedTo  string    `json:"assigned_to" binding:"required"`
	DueDate     time.Time `json:"due_date" binding:"required"`
}

type UpdateTaskRequest struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	Status      *string    `json:"status"`
	Priority    *string    `json:"priority"`
	AssignedTo  *string    `json:"assigned_to"`
	DueDate     *time.Time `json:"due_date"`
}

type TaskResponse struct {
	Task Task `json:"task"`
}

type TaskListResponse struct {
	Tasks      []Task `json:"tasks"`
	Pagination struct {
		CurrentPage int   `json:"current_page"`
		PageSize    int   `json:"page_size"`
		TotalItems  int64 `json:"total_items"`
		TotalPages  int   `json:"total_pages"`
	} `json:"pagination"`
}
