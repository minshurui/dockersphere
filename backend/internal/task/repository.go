package task

import "context"

// Repository defines the persistence interface for tasks.
type Repository interface {
	Save(ctx context.Context, t *Task) error
	FindByID(ctx context.Context, id string) (*Task, error)
	List(ctx context.Context) ([]*Task, error)
	Update(ctx context.Context, t *Task) error
}
