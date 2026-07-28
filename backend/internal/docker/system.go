package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	dockerclient "github.com/docker/docker/client"
)

// ImageService defines the interface for image operations.
type ImageService interface {
	List(ctx context.Context) ([]types.ImageSummary, error)
	Remove(ctx context.Context, id string) error
}

// SystemService defines the interface for Docker system info.
type SystemService interface {
	Info(ctx context.Context) (*types.Info, error)
	DiskUsage(ctx context.Context) (*types.DiskUsage, error)
}

type imageService struct {
	cli *dockerclient.Client
}

func NewImageService(cli *dockerclient.Client) ImageService {
	return &imageService{cli: cli}
}

func (s *imageService) List(ctx context.Context) ([]types.ImageSummary, error) {
	images, err := s.cli.ImageList(ctx, types.ImageListOptions{All: false})
	if err != nil {
		return nil, fmt.Errorf("list images: %w", err)
	}
	return images, nil
}

func (s *imageService) Remove(ctx context.Context, id string) error {
	_, err := s.cli.ImageRemove(ctx, id, types.ImageRemoveOptions{Force: false})
	if err != nil {
		return fmt.Errorf("remove image: %w", err)
	}
	return nil
}

type systemService struct {
	cli *dockerclient.Client
}

func NewSystemService(cli *dockerclient.Client) SystemService {
	return &systemService{cli: cli}
}

func (s *systemService) Info(ctx context.Context) (*types.Info, error) {
	info, err := s.cli.Info(ctx)
	if err != nil {
		return nil, fmt.Errorf("docker info: %w", err)
	}
	return &info, nil
}

func (s *systemService) DiskUsage(ctx context.Context) (*types.DiskUsage, error) {
	du, err := s.cli.DiskUsage(ctx, types.DiskUsageOptions{})
	if err != nil {
		return nil, fmt.Errorf("disk usage: %w", err)
	}
	return &du, nil
}
