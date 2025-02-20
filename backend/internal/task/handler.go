package task

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Handler struct {
	service  *Service
	logger   *zap.Logger
	upgrader websocket.Upgrader
}

func NewHandler(service *Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// Implement proper origin checking in production
				return true
			},
		},
	}
}

func (h *Handler) WebSocket(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("WebSocket upgrade failed", zap.Error(err))
		return
	}

	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	h.service.RegisterClient(conn)
	defer func() {
		h.service.UnregisterClient(conn)
		conn.Close()
	}()

	for {
		messageType, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Error("WebSocket read error", zap.Error(err))
			}
			break
		}

		// Reset read deadline after successful read
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		if messageType == websocket.PingMessage {
			if err := conn.WriteMessage(websocket.PongMessage, nil); err != nil {
				h.logger.Error("Failed to send pong", zap.Error(err))
				break
			}
		}
	}
}

func (h *Handler) CreateTask(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	resp, err := h.service.CreateTask(req, userID)
	if err != nil {
		h.logger.Error("Failed to create task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create task"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) UpdateTask(c *gin.Context) {
	taskID := c.Param("id")
	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	resp, err := h.service.UpdateTask(taskID, req, userID)
	if err != nil {
		if err == ErrTaskNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}
		h.logger.Error("Failed to update task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update task"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetTask(c *gin.Context) {
	taskID := c.Param("id")

	resp, err := h.service.GetTask(taskID)
	if err != nil {
		if err == ErrTaskNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}
		h.logger.Error("Failed to get task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get task"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ListTasks(c *gin.Context) {
	// Get filters from query parameters
	status := c.Query("status")
	assignedTo := c.Query("assigned_to")
	limit := 10 // Default limit

	resp, err := h.service.ListTasks(status, assignedTo, limit)
	if err != nil {
		h.logger.Error("Failed to list tasks", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tasks"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) DeleteTask(c *gin.Context) {
	taskID := c.Param("id")

	err := h.service.DeleteTask(taskID)
	if err != nil {
		if err == ErrTaskNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}
		h.logger.Error("Failed to delete task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "task deleted successfully"})
}

func (h *Handler) AssignTask(c *gin.Context) {
	taskID := c.Param("id")
	var req struct {
		AssignedTo string `json:"assigned_to" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.AssignTask(taskID, req.AssignedTo)
	if err != nil {
		if err == ErrTaskNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}
		h.logger.Error("Failed to assign task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to assign task"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
