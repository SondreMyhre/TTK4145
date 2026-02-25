package supervisor

import (
	"context"
)


type Worker interface {
	Run(ctx context.Context) error
}
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

