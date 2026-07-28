package docker

import (
	"context"

	"github.com/minshurui/dockersphere/internal/model"
)

// ContainerService defines the interface for container operations.
type ContainerService interface {
	List(ctx context.Context) ([]model.Container, error)
	Inspect(ctx context.Context, id string) (*model.Container, error)
	Stats(ctx context.Context, id string) (*model.ContainerStats, error)
	Logs(ctx context.Context, id string, tail int) ([]string, error)
	Exec(ctx context.Context, id string, cmd []string) (string, error)
	Start(ctx context.Context, id string) error
	Stop(ctx context.Context, id string) error
	Restart(ctx context.Context, id string) error
	Remove(ctx context.Context, id string) error
}

// HealthChecker defines the interface for health checks.
type HealthChecker interface {
	Check(ctx context.Context) error
}
