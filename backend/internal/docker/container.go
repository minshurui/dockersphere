package docker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"

	dockerclient "github.com/docker/docker/client"
	"github.com/yourname/dockersphere/internal/model"
)

type containerService struct {
	cli *dockerclient.Client
}

// NewContainerService creates a ContainerService backed by the Docker API.
func NewContainerService(cli *dockerclient.Client) ContainerService {
	return &containerService{cli: cli}
}

func (s *containerService) List(ctx context.Context) ([]model.Container, error) {
	listOpts := container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(),
	}
	containers, err := s.cli.ContainerList(ctx, listOpts)
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}

	result := make([]model.Container, 0, len(containers))
	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}
		ports := make([]model.ContainerPort, 0, len(c.Ports))
		for _, p := range c.Ports {
			ports = append(ports, model.ContainerPort{
				PrivatePort: p.PrivatePort,
				PublicPort:  p.PublicPort,
				Type:        p.Type,
				IP:          p.IP,
			})
		}
		result = append(result, model.Container{
			ID:      c.ID[:12],
			Name:    name,
			Image:   c.Image,
			State:   c.State,
			Status:  c.Status,
			Created: time.Unix(c.Created, 0),
			Ports:   ports,
			Labels:  c.Labels,
		})
	}
	return result, nil
}

func (s *containerService) Start(ctx context.Context, id string) error {
	return s.cli.ContainerStart(ctx, id, container.StartOptions{})
}

func (s *containerService) Stop(ctx context.Context, id string) error {
	return s.cli.ContainerStop(ctx, id, container.StopOptions{})
}

func (s *containerService) Restart(ctx context.Context, id string) error {
	return s.cli.ContainerRestart(ctx, id, container.StopOptions{})
}

func (s *containerService) Remove(ctx context.Context, id string) error {
	return s.cli.ContainerRemove(ctx, id, container.RemoveOptions{Force: true})
}
