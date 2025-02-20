package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Email     string         `gorm:"type:varchar(255);unique;not null;index" json:"email"`
	Password  string         `gorm:"type:varchar(255);not null" json:"-"`
	CreatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	AssignedTasks []Task `gorm:"foreignKey:AssignedTo;constraint:OnDelete:SET NULL" json:"assigned_tasks,omitempty"`
	CreatedTasks  []Task `gorm:"foreignKey:CreatedBy;constraint:OnDelete:SET NULL" json:"created_tasks,omitempty"`
}

type TaskStatus string
type TaskPriority string

const (
	StatusPending    TaskStatus = "pending"
	StatusInProgress TaskStatus = "in_progress"
	StatusCompleted  TaskStatus = "completed"

	PriorityLow    TaskPriority = "low"
	PriorityMedium TaskPriority = "medium"
	PriorityHigh   TaskPriority = "high"
)

type Task struct {
	ID          string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Title       string         `gorm:"type:varchar(255);not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Status      TaskStatus     `gorm:"type:varchar(50);not null;default:'pending';check:status IN ('pending', 'in_progress', 'completed')" json:"status"`
	Priority    TaskPriority   `gorm:"type:varchar(50);not null;check:priority IN ('low', 'medium', 'high')" json:"priority"`
	AssignedTo  string         `gorm:"type:uuid;index" json:"assigned_to"`
	CreatedBy   string         `gorm:"type:uuid;not null;index" json:"created_by"`
	CreatedAt   time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DueDate     time.Time      `gorm:"not null;index" json:"due_date"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	AssignedUser *User `gorm:"foreignKey:AssignedTo;references:ID" json:"assigned_user,omitempty"`
	Creator      *User `gorm:"foreignKey:CreatedBy;references:ID" json:"creator,omitempty"`
}
