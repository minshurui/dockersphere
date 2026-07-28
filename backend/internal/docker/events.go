package docker

import (
	"context"
	"log"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"

	"github.com/minshurui/dockersphere/internal/event"
)

// EventListener watches Docker daemon events and publishes them to the EventBus.
type EventListener struct {
	cli *dockerclient.Client
	bus *event.Bus
}

// NewEventListener creates a new EventListener.
func NewEventListener(cli *dockerclient.Client, bus *event.Bus) *EventListener {
	return &EventListener{cli: cli, bus: bus}
}

// Start begins listening for Docker events. It blocks until ctx is cancelled.
func (l *EventListener) Start(ctx context.Context) {
	const maxRetries = 10
	f := filters.NewArgs()
	f.Add("type", "container")

	for retry := 0; retry < maxRetries; retry++ {
		eventCh, errCh := l.cli.Events(ctx, types.EventsOptions{Filters: f})

	inner:
		for {
			select {
			case <-ctx.Done():
				log.Println("[EventListener] stopped")
				return
			case e, ok := <-eventCh:
				if !ok {
					break inner
				}
				l.handleEvent(e)
			case err, ok := <-errCh:
				if !ok {
					break inner
				}
				if ctx.Err() != nil {
					return
				}
				log.Printf("[EventListener] error: %v, reconnecting (retry %d/%d)...", err, retry+1, maxRetries)
				break inner
			}
		}

		select {
		case <-ctx.Done():
			log.Println("[EventListener] stopped")
			return
		default:
			time.Sleep(3 * time.Second)
		}
	}
	log.Println("[EventListener] max retries reached, giving up")
}

func (l *EventListener) handleEvent(e events.Message) {
	containerID := e.Actor.ID
	if len(containerID) > 12 {
		containerID = containerID[:12]
	}

	var eventType string
	switch e.Action {
	case "start":
		eventType = event.ContainerStarted
	case "stop":
		eventType = event.ContainerStopped
	case "die":
		eventType = event.ContainerStopped
	case "restart":
		eventType = event.ContainerRestarted
	case "create":
		eventType = event.ContainerCreated
	case "destroy":
		eventType = event.ContainerDestroyed
	case "pause":
		eventType = event.ContainerPaused
	case "unpause":
		eventType = event.ContainerUnpaused
	default:
		// Skip unhandled actions
		return
	}

	name := ""
	if n, ok := e.Actor.Attributes["name"]; ok {
		name = n
	}

	data := map[string]interface{}{
		"container_id": containerID,
		"name":         name,
		"image":        e.Actor.Attributes["image"],
		"action":       string(e.Action),
	}

	l.bus.Publish(eventType, data)
	log.Printf("[EventListener] %s: container=%s name=%s", eventType, containerID, name)
}
