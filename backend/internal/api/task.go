package api

import (
	"github.com/gin-gonic/gin"

	"github.com/minshurui/dockersphere/internal/model"
	"github.com/minshurui/dockersphere/internal/task"
)

// TaskHandler handles task-related HTTP requests.
type TaskHandler struct {
	service *task.Service
}

// NewTaskHandler creates a new TaskHandler.
func NewTaskHandler(service *task.Service) *TaskHandler {
	return &TaskHandler{service: service}
}

// List returns all tasks.
func (h *TaskHandler) List(c *gin.Context) {
	tasks, err := h.service.List(c.Request.Context())
	if err != nil {
		model.InternalError(c, err.Error())
		return
	}
	model.OK(c, tasks)
}

// Get returns a task by ID.
func (h *TaskHandler) Get(c *gin.Context) {
	id := c.Param("id")
	t, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		model.NotFound(c, err.Error())
		return
	}
	model.OK(c, t)
}
