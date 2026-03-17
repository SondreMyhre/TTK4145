package supervisor

import (
	"context"
	"sync"
	"time"
)

const (
	MaxRestarts  = 5
	MaxTime      = 5 * time.Second
	RestartDelay = 200 * time.Millisecond
)

// Worker interface for any runnable component
type Worker interface {
	Run(ctx context.Context) error
}

// WorkerFunc adapter - makes plain functions behave like a Worker
type WorkerFunc func(ctx context.Context) error

// RestartPolicy determines when a child should be restarted
type RestartPolicy int

const (
	Permanent RestartPolicy = iota // Always restart
	Transient                      // Restart only on error (not clean exit)
	Temporary                      // Never restart
)

// ChildSpec defines a supervised worker
type ChildSpec struct {
	Name    string
	Worker  Worker
	Restart RestartPolicy
}

// SupervisorConfig holds supervisor settings
type SupervisorConfig struct {
	MaxRestarts  int           // Max restarts allowed within MaxTime window
	MaxTime      time.Duration // Time window for MaxRestarts
	RestartDelay time.Duration // Delay between restarts
}

// Supervisor manages multiple children with restart capabilities
type Supervisor struct {
	Children []ChildSpec
	Config   SupervisorConfig
}

// childState tracks a running child (internal)
type childState struct {
	spec    ChildSpec
	cancel  context.CancelFunc
	tracker *restartTracker
	mu      sync.Mutex
}

// restartTracker tracks restart frequency for rate limiting (internal)
type restartTracker struct {
	timestamps []time.Time
	maxCount   int
	window     time.Duration
}
