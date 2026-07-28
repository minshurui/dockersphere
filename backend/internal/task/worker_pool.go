package task

import (
	"context"
	"log"
	"sync"
)

// WorkItem represents a unit of work to be executed by the pool.
type WorkItem struct {
	TaskID string
	Job    Job
}

// WorkerPool manages a fixed number of worker goroutines.
type WorkerPool struct {
	service   *Service
	workers   int
	workCh    chan WorkItem
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewWorkerPool creates a new worker pool with the given concurrency limit.
func NewWorkerPool(service *Service, workers int) *WorkerPool {
	if workers < 1 {
		workers = 1
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		service: service,
		workers: workers,
		workCh:  make(chan WorkItem, workers*2),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start launches the worker goroutines.
func (p *WorkerPool) Start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
	log.Printf("[WorkerPool] started with %d workers", p.workers)
}

// Submit adds a work item to the pool.
func (p *WorkerPool) Submit(item WorkItem) {
	p.workCh <- item
}

// Stop gracefully shuts down the pool.
func (p *WorkerPool) Stop() {
	close(p.workCh)
	p.cancel()
	p.wg.Wait()
	log.Println("[WorkerPool] stopped")
}

func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()
	for item := range p.workCh {
		log.Printf("[Worker %d] executing task %s", id, item.TaskID)
		p.service.Run(p.ctx, item.TaskID, item.Job)
	}
}
