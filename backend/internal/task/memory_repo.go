package task

import (
	"context"
	"fmt"
	"sync"
)

// MemoryRepo is an in-memory implementation of Repository.
type MemoryRepo struct {
	mu    sync.RWMutex
	tasks map[string]*Task
}

// NewMemoryRepo creates a new in-memory task repository.
func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{
		tasks: make(map[string]*Task),
	}
}

func (r *MemoryRepo) Save(_ context.Context, t *Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks[t.ID] = t
	return nil
}

func (r *MemoryRepo) FindByID(_ context.Context, id string) (*Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tasks[id]
	if !ok {
		return nil, fmt.Errorf("task %s not found", id)
	}
	return t, nil
}

func (r *MemoryRepo) List(_ context.Context) ([]*Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*Task, 0, len(r.tasks))
	for _, t := range r.tasks {
		result = append(result, t)
	}
	return result, nil
}

func (r *MemoryRepo) Update(_ context.Context, t *Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tasks[t.ID]; !ok {
		return fmt.Errorf("task %s not found", t.ID)
	}
	r.tasks[t.ID] = t
	return nil
}
