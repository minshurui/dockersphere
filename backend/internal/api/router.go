package api

import (
	"github.com/gin-gonic/gin"

	"github.com/minshurui/dockersphere/internal/audit"
	"github.com/minshurui/dockersphere/internal/docker"
	"github.com/minshurui/dockersphere/internal/middleware"
	"github.com/minshurui/dockersphere/internal/task"
	"github.com/minshurui/dockersphere/internal/ws"
)

// SetupRouter configures all API routes.
func SetupRouter(
	mode string,
	containerSvc docker.ContainerService,
	taskSvc *task.Service,
	pool *task.WorkerPool,
	hub *ws.Hub,
	auditStore *audit.Store,
) *gin.Engine {
	gin.SetMode(mode)
	r := gin.New()

	// Global middleware
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS())

	// Health checks
	r.GET("/health/live", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/health/ready", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	// WebSocket
	r.GET("/ws", func(c *gin.Context) {
		hub.ServeWS(c.Writer, c.Request)
	})

	// API v1
	v1 := r.Group("/api/v1")
	{
		containerHandler := NewContainerHandler(containerSvc, taskSvc, pool, auditStore)
		taskHandler := NewTaskHandler(taskSvc)

		v1.GET("/containers", containerHandler.List)
		v1.POST("/containers/:id/action", containerHandler.Action)

		v1.GET("/tasks", taskHandler.List)
		v1.GET("/tasks/:id", taskHandler.Get)

		if auditStore != nil {
			auditHandler := NewAuditHandler(auditStore)
			v1.GET("/audit", auditHandler.List)
		}
	}

	return r
}
