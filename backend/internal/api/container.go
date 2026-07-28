package api

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/yourname/dockersphere/internal/audit"
	"github.com/yourname/dockersphere/internal/docker"
	"github.com/yourname/dockersphere/internal/model"
	"github.com/yourname/dockersphere/internal/task"
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

// Action performs an async action on a container.
func (h *ContainerHandler) Action(c *gin.Context) {
	id := c.Param("id")
	var req model.ContainerAction
	if err := c.ShouldBindJSON(&req); err != nil {
		model.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	ctx := c.Request.Context()
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

func (h *ContainerHandler) buildJob(ctx context.Context, action, id string) task.Job {
	return func() error {
		switch action {
		case "start":
			return h.service.Start(ctx, id)
		case "stop":
			return h.service.Stop(ctx, id)
		case "restart":
			return h.service.Restart(ctx, id)
		case "remove":
			return h.service.Remove(ctx, id)
		default:
			return nil
		}
	}
}
