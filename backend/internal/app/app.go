package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/minshurui/dockersphere/internal/audit"
	"github.com/minshurui/dockersphere/internal/api"
	"github.com/minshurui/dockersphere/internal/config"
	"github.com/minshurui/dockersphere/internal/docker"
	"github.com/minshurui/dockersphere/internal/event"
	"github.com/minshurui/dockersphere/internal/task"
	"github.com/minshurui/dockersphere/internal/ws"
)

// App holds all application dependencies and lifecycle.
type App struct {
	cfg          *config.Config
	containerSvc docker.ContainerService
	healthCheck  docker.HealthChecker
	bus          *event.Bus
	taskSvc      *task.Service
	pool         *task.WorkerPool
	hub          *ws.Hub
	auditStore   *audit.Store
	listener     *docker.EventListener
	imageSvc     docker.ImageService
	systemSvc    docker.SystemService
}

// New creates and initializes a new App.
func New(cfg *config.Config) (*App, error) {
	app := &App{cfg: cfg}

	// Ensure data directory exists before DB initialization
	if err := os.MkdirAll("data", 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	// Docker client
	cli, err := docker.NewClient(cfg.Docker.Host)
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}
	app.containerSvc = docker.NewContainerService(cli)
	app.healthCheck = docker.NewHealthChecker(cli)

	app.imageSvc = docker.NewImageService(cli)
	app.systemSvc = docker.NewSystemService(cli)

	// Event bus
	app.bus = event.NewBus()

	// Task repository
	var taskRepo task.Repository
	switch cfg.Task.Repository {
	case "sqlite":
		repo, err := task.NewSQLiteRepo("data/tasks.db")
		if err != nil {
			return nil, fmt.Errorf("create sqlite task repo: %w", err)
		}
		taskRepo = repo
	default:
		taskRepo = task.NewMemoryRepo()
	}

	// Task service & pool
	app.taskSvc = task.NewService(taskRepo, app.bus)
	app.pool = task.NewWorkerPool(app.taskSvc, cfg.Task.WorkerPoolSize)

	// WebSocket hub
	app.hub = ws.NewHub()

	// Wire bus -> hub (broadcast all events to WebSocket clients)
	app.bus.Subscribe(event.Wildcard, func(eventType string, data interface{}) {
		app.hub.Broadcast(eventType, data)
	})

	// Audit store
	if cfg.Audit.Enabled {
		store, err := audit.NewStore(cfg.Audit.DBPath)
		if err != nil {
			return nil, fmt.Errorf("create audit store: %w", err)
		}
		app.auditStore = store
	}

	// Docker event listener
	app.listener = docker.NewEventListener(cli, app.bus)

	return app, nil
}

// Run starts the application and blocks until shutdown.
func (a *App) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start worker pool
	a.pool.Start()
	defer a.pool.Stop()

	// Start WebSocket hub
	go a.hub.Run()

	// Start Docker event listener
	go a.listener.Start(ctx)

	// Setup router
	router := api.SetupRouter(
		a.cfg.Server.Mode,
		a.containerSvc,
		a.taskSvc,
		a.pool,
		a.hub,
		a.auditStore,
		a.imageSvc,
		a.systemSvc,
	)

	// Start HTTP server
	addr := fmt.Sprintf(":%d", a.cfg.Server.Port)
	log.Printf("[App] %s v%s starting on %s", a.cfg.App.Name, a.cfg.App.Version, addr)

	go func() {
		if err := router.Run(addr); err != nil {
			log.Fatalf("[App] server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[App] shutting down...")

	return nil
}
