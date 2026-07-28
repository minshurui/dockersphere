package event

import (
	"sync"
	"testing"
	"time"
)

func TestBus_SubscribeAndPublish(t *testing.T) {
	bus := NewBus()
	var received string
	var wg sync.WaitGroup
	wg.Add(1)

	bus.Subscribe("test.event", func(eventType string, data interface{}) {
		received = eventType
		wg.Done()
	})

	bus.Publish("test.event", nil)
	wg.Wait()

	if received != "test.event" {
		t.Errorf("expected test.event, got %s", received)
	}
}

func TestBus_Wildcard(t *testing.T) {
	bus := NewBus()
	count := 0
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(3)

	bus.Subscribe(Wildcard, func(eventType string, data interface{}) {
		mu.Lock()
		count++
		mu.Unlock()
		wg.Done()
	})

	bus.Publish("event1", nil)
	bus.Publish("event2", nil)
	bus.Publish("event3", nil)
	wg.Wait()

	if count != 3 {
		t.Errorf("expected 3 events, got %d", count)
	}
}

func TestBus_PatternMatching(t *testing.T) {
	bus := NewBus()
	count := 0
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(2)

	bus.Subscribe("container.*", func(eventType string, data interface{}) {
		mu.Lock()
		count++
		mu.Unlock()
		wg.Done()
	})

	bus.Publish("container.started", nil)
	bus.Publish("container.stopped", nil)
	bus.Publish("task.created", nil) // Should not match
	wg.Wait()

	time.Sleep(10 * time.Millisecond) // Allow async handlers to complete

	if count != 2 {
		t.Errorf("expected 2 container events, got %d", count)
	}
}
