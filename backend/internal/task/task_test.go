package task

import (
	"context"
	"testing"
	"time"

	"github.com/minshurui/dockersphere/internal/event"
)

func TestMemoryRepo_SaveAndFind(t *testing.T) {
	repo := NewMemoryRepo()
	ctx := context.Background()

	task := &Task{
		ID:     "test-1",
		Action: "start",
		Target: "container-1",
		Status: StatusPending,
	}

	if err := repo.Save(ctx, task); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	found, err := repo.FindByID(ctx, "test-1")
	if err != nil {
		t.Fatalf("find failed: %v", err)
	}

	if found.ID != task.ID {
		t.Errorf("expected ID %s, got %s", task.ID, found.ID)
	}
}

func TestMemoryRepo_List(t *testing.T) {
	repo := NewMemoryRepo()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		task := &Task{
			ID:     "task-" + string(rune('0'+i)),
			Action: "start",
			Target: "container",
			Status: StatusPending,
		}
		_ = repo.Save(ctx, task)
	}

	tasks, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}

	if len(tasks) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(tasks))
	}
}

func TestService_SubmitAndRun(t *testing.T) {
	repo := NewMemoryRepo()
	bus := event.NewBus()
	svc := NewService(repo, bus)
	ctx := context.Background()

	taskID, err := svc.Submit(ctx, "start", "container-1")
	if err != nil {
		t.Fatalf("submit failed: %v", err)
	}

	if taskID == "" {
		t.Error("expected non-empty task ID")
	}

	svc.Run(ctx, taskID, func() error {
		return nil
	})

	task, err := svc.Get(ctx, taskID)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}

	if task.Status != StatusCompleted {
		t.Errorf("expected status %s, got %s", StatusCompleted, task.Status)
	}
}

func TestWorkerPool(t *testing.T) {
	repo := NewMemoryRepo()
	bus := event.NewBus()
	svc := NewService(repo, bus)

	pool := NewWorkerPool(svc, 2)
	pool.Start()
	defer pool.Stop()

	ctx := context.Background()
	taskID, _ := svc.Submit(ctx, "test", "target")

	pool.Submit(WorkItem{
		TaskID: taskID,
		Job:    func() error { return nil },
	})

	// Wait for completion
	for i := 0; i < 10; i++ {
		task, _ := svc.Get(ctx, taskID)
		if task.Status == StatusCompleted {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Error("task did not complete in time")
}
