package supervisor

import (
	"context"
	"time"
)

// Worker works with all run(ctx)error
type Worker interface {
	Run(ctx context.Context) error
}

// Adapter, makes plain functions bahave like a worker
type WorkerFunc func(ctx context.Context) error

type RestartPolicy int

type ChildSpec struct {
	Name    string
	Worker  Worker
	Restart RestartPolicy
}

const(
	Permanent RestartPolicy = iota
	Transient
	Temporary
)

type Supervisor struct {
	Child        ChildSpec
	RestartDelay time.Duration
}

const (
	RestartDelay = 100 * time.Millisecond
)
