package supervisor

import (
	"context"
	"fmt"
)

func (fn WorkerFunc) Run(ctx context.Context) error{ return fn(ctx) }

//creates new child	
func New(child ChildSpec) Supervisor { 
	return Supervisor{
		Child:        child,
		RestartDelay: RestartDelay,
	}
}

// panic safe
func (supervisor Supervisor) runWorker(ctx context.Context) (err error) {
	defer func() {
		recovered := recover() 
		if recovered != nil {
			err = fmt.Errorf("worker panic: %v", recovered)
		}
	}()
	return supervisor.Child.Worker.Run(ctx)
}
