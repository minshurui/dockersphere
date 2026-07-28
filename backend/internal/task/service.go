package task

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourname/dockersphere/internal/event"
)

// Service manages task creation and execution.
type Service struct {
	repo Repository
	bus  *event.Bus
}

// NewService creates a new task service with the given repository and event bus.
func NewService(repo Repository, bus *event.Bus) *Service {
	return &Service{repo: repo, bus: bus}
}

// Submit creates a new task and returns its ID.
func (s *Service) Submit(ctx context.Context, action, target string) (string, error) {
	t := &Task{
		ID:        uuid.New().String(),
		Action:    action,
		Target:    target,
		Status:    StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.repo.Save(ctx, t); err != nil {
		return "", fmt.Errorf("save task: %w", err)
	}
	s.bus.Publish(event.TaskCreated, t)
	return t.ID, nil
}

// Run executes a job for the given task, updating its status.
func (s *Service) Run(ctx context.Context, taskID string, job Job) {
	t, err := s.repo.FindByID(ctx, taskID)
	if err != nil {
		return
	}
	t.Status = StatusRunning
	t.UpdatedAt = time.Now()
	_ = s.repo.Update(ctx, t)
	s.bus.Publish(event.TaskRunning, t)

	if err := job(); err != nil {
		t.Status = StatusFailed
		t.Err = err.Error()
		t.UpdatedAt = time.Now()
		_ = s.repo.Update(ctx, t)
		s.bus.Publish(event.TaskFailed, t)
		return
	}

	t.Status = StatusCompleted
	t.Result = "ok"
	t.UpdatedAt = time.Now()
	_ = s.repo.Update(ctx, t)
	s.bus.Publish(event.TaskCompleted, t)
}

// Get retrieves a task by ID.
func (s *Service) Get(ctx context.Context, id string) (*Task, error) {
	return s.repo.FindByID(ctx, id)
}

// List returns all tasks.
func (s *Service) List(ctx context.Context) ([]*Task, error) {
	return s.repo.List(ctx)
}
