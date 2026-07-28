package api

import (
	"net/http"

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
	imageSvc docker.ImageService,
	systemSvc docker.SystemService,
) *gin.Engine {
	gin.SetMode(mode)
	r := gin.New()

	// Trust only local networks
	_ = r.SetTrustedProxies([]string{"127.0.0.0/8", "::1", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"})

	// Global middleware
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS())

	// Health checks
	r.GET("/health/live", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/health/ready", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	// WebSocket (with optional token auth via query param)
	r.GET("/ws", func(c *gin.Context) {
		token := c.Query("token")
		if token != "" && token != "dockersphere" {
			c.JSON(401, gin.H{"code": 401, "message": "unauthorized"})
			return
		}
		hub.ServeWS(c.Writer, c.Request)
	})

	// API v1
	v1 := r.Group("/api/v1")
	{
		containerHandler := NewContainerHandler(containerSvc, taskSvc, pool, auditStore)
		taskHandler := NewTaskHandler(taskSvc)

		v1.GET("/containers", containerHandler.List)
		v1.GET("/containers/:id/stats", containerHandler.Stats)
		v1.GET("/containers/:id/logs", containerHandler.Logs)
		v1.POST("/containers/:id/exec", containerHandler.Exec)
		v1.POST("/containers/:id/action", containerHandler.Action)

		v1.GET("/tasks", taskHandler.List)
		v1.GET("/tasks/:id", taskHandler.Get)

		if auditStore != nil {
			auditHandler := NewAuditHandler(auditStore)
			v1.GET("/audit", auditHandler.List)
		}
	}

	// Compose files
	composeHandler := NewComposeHandler()
	v1.GET("/compose/files", composeHandler.List)
	v1.GET("/compose/file", composeHandler.Read)
	v1.PUT("/compose/file", composeHandler.Update)

	// System & Images
	systemHandler := NewSystemHandler(imageSvc, systemSvc)
	v1.GET("/images", systemHandler.Images)
	v1.DELETE("/images/:id", systemHandler.ImageRemove)
	v1.GET("/system/info", systemHandler.Info)
	v1.GET("/system/df", systemHandler.DiskUsage)

	// Serve built frontend SPA
	r.Static("/assets", "../frontend/dist/assets")
	r.NoRoute(func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "../frontend/dist/index.html")
	})

	return r
}
