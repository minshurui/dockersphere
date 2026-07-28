package api

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/minshurui/dockersphere/internal/audit"
	"github.com/minshurui/dockersphere/internal/docker"
	"github.com/minshurui/dockersphere/internal/model"
	"github.com/minshurui/dockersphere/internal/task"
)

// ContainerHandler handles container-related HTTP requests.
type ContainerHandler struct {
	service   docker.ContainerService
	taskSvc   *task.Service
	pool      *task.WorkerPool
	auditStore *audit.Store
}

// NewContainerHandler creates a new ContainerHandler.
func NewContainerHandler(
	service docker.ContainerService,
	taskSvc *task.Service,
	pool *task.WorkerPool,
	auditStore *audit.Store,
) *ContainerHandler {
	return &ContainerHandler{
		service:   service,
		taskSvc:   taskSvc,
		pool:      pool,
		auditStore: auditStore,
	}
}

// List returns all containers.
func (h *ContainerHandler) List(c *gin.Context) {
	containers, err := h.service.List(c.Request.Context())
	if err != nil {
		model.InternalError(c, err.Error())
		return
	}
	model.OK(c, containers)
}

// Stats returns real-time CPU/memory stats for a container.
func (h *ContainerHandler) Stats(c *gin.Context) {
	id := c.Param("id")
	stats, err := h.service.Stats(c.Request.Context(), id)
	if err != nil {
		model.NotFound(c, "container not found: "+id)
		return
	}
	model.OK(c, stats)
}

// Logs returns container logs.
func (h *ContainerHandler) Logs(c *gin.Context) {
	id := c.Param("id")
	tail := 100
	if t := c.Query("tail"); t != "" {
		if n, err := strconv.Atoi(t); err == nil && n > 0 {
			tail = n
		}
	}
	logs, err := h.service.Logs(c.Request.Context(), id, tail)
	if err != nil {
		model.NotFound(c, "container not found: "+id)
		return
	}
	model.OK(c, logs)
}

type execRequest struct {
	Cmd  string   `json:"cmd"`
	Args []string `json:"args"`
}

// Exec runs a command in a container.
func (h *ContainerHandler) Exec(c *gin.Context) {
	id := c.Param("id")
	var req execRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		model.BadRequest(c, "invalid request")
		return
	}
	cmd := []string{req.Cmd}
	cmd = append(cmd, req.Args...)
	output, err := h.service.Exec(c.Request.Context(), id, cmd)
	if err != nil {
		model.InternalError(c, err.Error())
		return
	}
	model.OK(c, output)
}

// Action performs an async action on a container.
func (h *ContainerHandler) Action(c *gin.Context) {
	id := c.Param("id")
	var req model.ContainerAction
	if err := c.ShouldBindJSON(&req); err != nil {
		model.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	ctx := c.Request.Context()

	// Verify container exists before submitting the task
	if _, err := h.service.Inspect(ctx, id); err != nil {
		model.NotFound(c, "container not found: "+id)
		return
	}

	taskID, err := h.taskSvc.Submit(ctx, req.Action, id)
	if err != nil {
		model.InternalError(c, err.Error())
		return
	}

	// Record audit
	if h.auditStore != nil {
		_ = h.auditStore.Record(ctx, req.Action, id, "system", "")
	}

	// Submit to worker pool
	h.pool.Submit(task.WorkItem{
		TaskID: taskID,
		Job:    h.buildJob(ctx, req.Action, id),
	})

	model.Accepted(c, gin.H{"task_id": taskID})
}

func (h *ContainerHandler) buildJob(_ context.Context, action, id string) task.Job {
	bg := context.Background()
	return func() error {
		switch action {
		case "start":
			return h.service.Start(bg, id)
		case "stop":
			return h.service.Stop(bg, id)
		case "restart":
			return h.service.Restart(bg, id)
		case "remove":
			return h.service.Remove(bg, id)
		default:
			return nil
		}
	}
}
