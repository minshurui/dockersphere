package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
)

// NewClient creates a new Docker API client from the given host.
func NewClient(host string) (*client.Client, error) {
	opts := []client.Opt{
		client.WithAPIVersionNegotiation(),
	}
	if host != "" {
		opts = append(opts, client.WithHost(host))
	}
	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}
	// Verify connectivity
	ctx := context.Background()
	if _, err := cli.Ping(ctx); err != nil {
		cli.Close()
		return nil, fmt.Errorf("docker ping: %w", err)
	}
	return cli, nil
}
