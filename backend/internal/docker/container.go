package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"

	dockerclient "github.com/docker/docker/client"
	"github.com/minshurui/dockersphere/internal/model"
)

func toContainer(c types.Container) model.Container {
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
	return model.Container{
		ID:      c.ID[:12],
		Name:    name,
		Image:   c.Image,
		State:   c.State,
		Status:  c.Status,
		Created: time.Unix(c.Created, 0),
		Ports:   ports,
		Labels:  c.Labels,
	}
}

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
		result = append(result, toContainer(c))
	}
	return result, nil
}

func (s *containerService) Inspect(ctx context.Context, id string) (*model.Container, error) {
	list, err := s.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}
	for _, c := range list {
		name := strings.TrimPrefix(c.Names[0], "/")
		if name == id || c.ID == id || strings.HasPrefix(c.ID, id) {
			result := toContainer(c)
			return &result, nil
		}
	}
	return nil, fmt.Errorf("container %s not found", id)
}

func (s *containerService) Stats(ctx context.Context, id string) (*model.ContainerStats, error) {
	resp, err := s.cli.ContainerStats(ctx, id, false)
	if err != nil {
		return nil, fmt.Errorf("container stats: %w", err)
	}
	defer resp.Body.Close()

	var v *model.ContainerStats
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("decode stats: %w", err)
	}
	return v, nil
}

func (s *containerService) Logs(ctx context.Context, id string, tail int) ([]string, error) {
	if tail <= 0 || tail > 500 {
		tail = 100
	}
	opts := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       strconv.Itoa(tail),
	}
	reader, err := s.cli.ContainerLogs(ctx, id, opts)
	if err != nil {
		return nil, fmt.Errorf("container logs: %w", err)
	}
	defer reader.Close()

	var lines []string
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		// Strip Docker log header (8 bytes)
		if len(line) > 8 {
			line = line[8:]
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan logs: %w", err)
	}
	return lines, nil
}

func (s *containerService) Exec(ctx context.Context, id string, cmd []string) (string, error) {
	if len(cmd) == 0 {
		cmd = []string{"sh", "-c", "echo hello"}
	}
	execCfg := types.ExecConfig{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	}
	execID, err := s.cli.ContainerExecCreate(ctx, id, execCfg)
	if err != nil {
		return "", fmt.Errorf("exec create: %w", err)
	}

	resp, err := s.cli.ContainerExecAttach(ctx, execID.ID, types.ExecStartCheck{})
	if err != nil {
		return "", fmt.Errorf("exec attach: %w", err)
	}
	defer resp.Close()

	output, err := io.ReadAll(resp.Reader)
	if err != nil {
		return "", fmt.Errorf("exec read: %w", err)
	}
	return string(output), nil
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
