package task

import "time"

// Status represents the state of a task.
type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

// Task represents an asynchronous operation.
type Task struct {
	ID        string      `json:"id"`
	Action    string      `json:"action"`
	Target    string      `json:"target"`
	Status    Status      `json:"status"`
	Result    string      `json:"result,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	Err       string      `json:"error,omitempty"`
}

// Job is a function that performs work.
type Job func() error
