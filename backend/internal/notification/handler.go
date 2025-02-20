package notification

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	service *Service
	logger  *zap.Logger
}

func NewHandler(service *Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) HandleTaskEvent(c *gin.Context) {
	var event NotificationEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		h.logger.Error("Invalid notification event", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate event data
	if event.Task.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task data"})
		return
	}

	// Send notification asynchronously
	go func() {
		h.service.SendNotification(event)
	}()

	c.JSON(http.StatusAccepted, gin.H{"message": "notification queued"})
}
