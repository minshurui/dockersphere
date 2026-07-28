package event

import (
	"log"
	"strings"
	"sync"
)

// Handler is a callback invoked when an event is received.
type Handler func(eventType string, data interface{})

// Bus is an asynchronous event bus supporting wildcard subscriptions.
type Bus struct {
	mu          sync.RWMutex
	subscribers map[string][]Handler
}

// NewBus creates a new event bus.
func NewBus() *Bus {
	return &Bus{
		subscribers: make(map[string][]Handler),
	}
}

// Subscribe registers a handler for a specific event type.
// Use Wildcard ("*") to subscribe to all events.
func (b *Bus) Subscribe(eventType string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[eventType] = append(b.subscribers[eventType], handler)
}

// Publish dispatches an event to all matching subscribers asynchronously.
func (b *Bus) Publish(eventType string, data interface{}) {
	b.mu.RLock()
	handlers := make([]Handler, 0)

	// Exact match
	if hs, ok := b.subscribers[eventType]; ok {
		handlers = append(handlers, hs...)
	}

	// Wildcard subscribers receive all events
	if eventType != Wildcard {
		if hs, ok := b.subscribers[Wildcard]; ok {
			handlers = append(handlers, hs...)
		}
	}

	// Pattern matching: e.g. "container.*" matches "container.started"
	for pattern, hs := range b.subscribers {
		if pattern == eventType || pattern == Wildcard {
			continue
		}
		if matchPattern(pattern, eventType) {
			handlers = append(handlers, hs...)
		}
	}
	b.mu.RUnlock()

	for _, h := range handlers {
		go func(fn Handler) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[EventBus] handler panic: %v", r)
				}
			}()
			fn(eventType, data)
		}(h)
	}
}

// matchPattern checks if an event type matches a pattern.
// Supports "prefix.*" style patterns.
func matchPattern(pattern, eventType string) bool {
	if !strings.Contains(pattern, "*") {
		return pattern == eventType
	}
	prefix := strings.TrimSuffix(pattern, "*")
	return strings.HasPrefix(eventType, prefix)
}
