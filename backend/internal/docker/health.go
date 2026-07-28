package docker

import (
	"context"
	"fmt"

	dockerclient "github.com/docker/docker/client"
)

type healthChecker struct {
	cli *dockerclient.Client
}

// NewHealthChecker creates a HealthChecker that pings the Docker daemon.
func NewHealthChecker(cli *dockerclient.Client) HealthChecker {
	return &healthChecker{cli: cli}
}

func (h *healthChecker) Check(ctx context.Context) error {
	_, err := h.cli.Ping(ctx)
	if err != nil {
		return fmt.Errorf("docker daemon unreachable: %w", err)
	}
	return nil
}
