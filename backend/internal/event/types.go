package event

// Event type constants for the DockerSphere event system.
const (
	// Container lifecycle events
	ContainerCreated    = "container.created"
	ContainerStarted    = "container.started"
	ContainerStopped    = "container.stopped"
	ContainerRestarted  = "container.restarted"
	ContainerDestroyed  = "container.destroyed"
	ContainerPaused     = "container.paused"
	ContainerUnpaused   = "container.unpaused"

	// Task events
	TaskCreated   = "task.created"
	TaskRunning   = "task.running"
	TaskCompleted = "task.completed"
	TaskFailed    = "task.failed"

	// Audit events
	AuditRecorded = "audit.recorded"

	// Wildcard matches any event
	Wildcard = "*"
)
