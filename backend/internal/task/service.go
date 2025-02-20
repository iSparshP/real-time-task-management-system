package task

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/iSparshP/real-time-task-management-system/internal/common"
	"github.com/iSparshP/real-time-task-management-system/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	PriorityLow    = models.PriorityLow
	PriorityMedium = models.PriorityMedium
	PriorityHigh   = models.PriorityHigh
)

const (
	StatusPending    = models.StatusPending
	StatusInProgress = models.StatusInProgress
	StatusCompleted  = models.StatusCompleted
)

type Service struct {
	db         *gorm.DB
	clients    map[*websocket.Conn]*sync.Mutex // Change to mutex per client
	broadcast  chan WebSocketMessage           // Change to typed channel
	clientsMux sync.RWMutex
	logger     *zap.Logger
}

func NewService(db *gorm.DB, logger *zap.Logger) *Service {
	s := &Service{
		db:        db,
		clients:   make(map[*websocket.Conn]*sync.Mutex),
		broadcast: make(chan WebSocketMessage),
		logger:    logger,
	}
	go s.handleBroadcast()
	return s
}

func (s *Service) handleBroadcast() {
	for msg := range s.broadcast {
		s.clientsMux.RLock()
		for client, mutex := range s.clients {
			go func(c *websocket.Conn, m *sync.Mutex) {
				m.Lock()
				defer m.Unlock()
				if err := c.WriteJSON(msg); err != nil {
					s.logger.Error("Failed to send message", zap.Error(err))
					s.UnregisterClient(c)
				}
			}(client, mutex)
		}
		s.clientsMux.RUnlock()
	}
}

func (s *Service) RegisterClient(conn *websocket.Conn) {
	s.clientsMux.Lock()
	s.clients[conn] = &sync.Mutex{}
	s.clientsMux.Unlock()
}

func (s *Service) UnregisterClient(conn *websocket.Conn) {
	s.clientsMux.Lock()
	delete(s.clients, conn)
	s.clientsMux.Unlock()
}

func (s *Service) CreateTask(req CreateTaskRequest, userID string) (*TaskResponse, error) {
	task := &Task{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Description: req.Description,
		Status:      models.StatusPending,
		Priority:    models.TaskPriority(req.Priority),
		AssignedTo:  req.AssignedTo,
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		DueDate:     req.DueDate,
	}

	if err := s.validateTask(task); err != nil {
		return nil, err
	}

	if err := s.db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	s.broadcast <- WebSocketMessage{
		Type:    MessageTypeTaskCreated,
		Payload: *task,
	}
	return &TaskResponse{Task: *task}, nil
}

func (s *Service) canModifyTask(userID string, task *Task) bool {
	return task.CreatedBy == userID || task.AssignedTo == userID
}

func (s *Service) UpdateTask(taskID string, req UpdateTaskRequest, userID string) (*TaskResponse, error) {
	var task Task
	if err := s.db.First(&task, "id = ?", taskID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	if !s.canModifyTask(userID, &task) {
		return nil, ErrUnauthorized
	}

	// Apply updates
	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Status != nil {
		task.Status = models.TaskStatus(*req.Status)
	}
	if req.Priority != nil {
		task.Priority = models.TaskPriority(*req.Priority)
	}
	if req.AssignedTo != nil {
		task.AssignedTo = *req.AssignedTo
	}
	if req.DueDate != nil {
		task.DueDate = *req.DueDate
	}
	task.UpdatedAt = time.Now()

	// Validate updated task
	if err := s.validateTask(&task); err != nil {
		return nil, err
	}

	if err := s.db.Save(&task).Error; err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	s.broadcast <- WebSocketMessage{
		Type:    MessageTypeTaskUpdated,
		Payload: task,
	}
	return &TaskResponse{Task: task}, nil
}

func (s *Service) GetTask(taskID string) (*TaskResponse, error) {
	task := &Task{}
	if err := s.db.First(task, "id = ?", taskID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}
	return &TaskResponse{Task: *task}, nil
}

func (s *Service) ListTasks(status string, assignedTo string, page int) (*TaskListResponse, error) {
	var tasks []Task
	query := s.db

	if status != "" {
		if !isValidStatus(models.TaskStatus(status)) {
			return nil, ErrInvalidStatus
		}
		query = query.Where("status = ?", status)
	}

	if assignedTo != "" {
		query = query.Where("assigned_to = ?", assignedTo)
	}

	offset := (page - 1) * common.AppConfig.TaskPageSize
	query = query.Offset(offset).Limit(common.AppConfig.TaskPageSize)

	if err := query.Order("created_at desc").Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	return &TaskListResponse{Tasks: tasks}, nil
}

func (s *Service) ListTasksWithFilters(filter TaskFilter, pagination PaginationParams, sort SortParams) (*TaskListResponse, error) {
	var tasks []Task
	query := s.db.Model(&Task{})

	// Apply filters
	if filter.Status != nil {
		if !isValidStatus(TaskStatus(*filter.Status)) {
			return nil, ErrInvalidStatus
		}
		query = query.Where("status = ?", *filter.Status)
	}

	if filter.Priority != nil {
		if !isValidPriority(TaskPriority(*filter.Priority)) {
			return nil, ErrInvalidPriority
		}
		query = query.Where("priority = ?", *filter.Priority)
	}

	if filter.AssignedTo != nil {
		query = query.Where("assigned_to = ?", *filter.AssignedTo)
	}

	if filter.CreatedBy != nil {
		query = query.Where("created_by = ?", *filter.CreatedBy)
	}

	if filter.DueBefore != nil {
		query = query.Where("due_date <= ?", *filter.DueBefore)
	}

	if filter.DueAfter != nil {
		query = query.Where("due_date >= ?", *filter.DueAfter)
	}

	// Apply sorting
	sortOrder := "DESC"
	if sort.SortOrder == "asc" {
		sortOrder = "ASC"
	}
	query = query.Order(fmt.Sprintf("%s %s", sort.SortBy, sortOrder))

	// Apply pagination
	offset := (pagination.Page - 1) * pagination.PageSize
	query = query.Offset(offset).Limit(pagination.PageSize)

	// Execute query
	if err := query.Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	// Get total count for pagination
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	return &TaskListResponse{
		Tasks: tasks,
		Pagination: struct {
			CurrentPage int   `json:"current_page"`
			PageSize    int   `json:"page_size"`
			TotalItems  int64 `json:"total_items"`
			TotalPages  int   `json:"total_pages"`
		}{
			CurrentPage: pagination.Page,
			PageSize:    pagination.PageSize,
			TotalItems:  total,
			TotalPages:  int(math.Ceil(float64(total) / float64(pagination.PageSize))),
		},
	}, nil
}

func (s *Service) DeleteTask(taskID string) error {
	result := s.db.Delete(&Task{}, "id = ?", taskID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete task: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrTaskNotFound
	}

	s.broadcast <- WebSocketMessage{
		Type: MessageTypeTaskDeleted,
		Payload: Task{
			ID:     taskID,
			Status: "deleted",
		},
	}
	return nil
}

func (s *Service) AssignTask(taskID string, assignedTo string) (*TaskResponse, error) {
	task := &Task{}
	if err := s.db.First(task, "id = ?", taskID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	task.AssignedTo = assignedTo
	task.UpdatedAt = time.Now()

	if err := s.validateTask(task); err != nil {
		return nil, err
	}

	if err := s.db.Save(task).Error; err != nil {
		return nil, fmt.Errorf("failed to assign task: %w", err)
	}

	s.broadcast <- WebSocketMessage{
		Type:    MessageTypeTaskUpdated,
		Payload: *task,
	}
	return &TaskResponse{Task: *task}, nil
}

func isValidStatus(status models.TaskStatus) bool {
	validStatuses := []models.TaskStatus{
		models.StatusPending,
		models.StatusInProgress,
		models.StatusCompleted,
	}
	for _, s := range validStatuses {
		if status == s {
			return true
		}
	}
	return false
}

func isValidPriority(priority models.TaskPriority) bool {
	validPriorities := []models.TaskPriority{
		models.PriorityLow,
		models.PriorityMedium,
		models.PriorityHigh,
	}
	for _, p := range validPriorities {
		if priority == p {
			return true
		}
	}
	return false
}

func isValidDueDate(dueDate time.Time) bool {
	return !dueDate.Before(time.Now())
}

func (s *Service) validateTaskCreate(task *Task) error {
	if task.Title == "" {
		return errors.New("title is required")
	}
	if len(task.Description) > common.AppConfig.TaskMaxDescLength {
		return fmt.Errorf("description exceeds maximum length of %d", common.AppConfig.TaskMaxDescLength)
	}
	if task.DueDate.Before(time.Now()) {
		return ErrInvalidDueDate
	}
	if !isValidPriority(task.Priority) {
		return ErrInvalidPriority
	}
	return nil
}

func (s *Service) validateTask(task *Task) error {
	// Title validation
	if task.Title == "" {
		return fmt.Errorf("title is required")
	}
	if len(task.Title) > 255 {
		return fmt.Errorf("title exceeds maximum length of 255 characters")
	}

	// Description validation
	maxDescLen := common.AppConfig.TaskMaxDescLength
	if maxDescLen <= 0 {
		maxDescLen = 1000 // Fallback default
	}
	if len(task.Description) > maxDescLen {
		return fmt.Errorf("description exceeds maximum length of %d characters", maxDescLen)
	}

	// Status validation
	if task.Status != "" && !isValidStatus(task.Status) {
		return ErrInvalidStatus
	}

	// Priority validation
	if !isValidPriority(task.Priority) {
		return ErrInvalidPriority
	}

	// Due date validation
	if !task.DueDate.IsZero() && task.DueDate.Before(time.Now()) {
		return ErrInvalidDueDate
	}

	// AssignedTo validation
	if task.AssignedTo != "" {
		var user models.User
		if err := s.db.First(&user, "id = ?", task.AssignedTo).Error; err != nil {
			return ErrInvalidAssignment
		}
	}

	return nil
}
