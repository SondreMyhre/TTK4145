package supervisor

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Run starts and supervises all children
func (sup Supervisor) Run(ctx context.Context) error {
	if len(sup.Children) == 0 {
		return fmt.Errorf("supervisor: no children to supervise")
	}

	// Initialize child states
	states := make([]*childState, len(sup.Children))
	for index, child := range sup.Children {
		states[index] = &childState{
			spec:    child,
			tracker: newRestartTracker(sup.Config.MaxRestarts, sup.Config.MaxTime),
		}
	}

	var waitgroup sync.WaitGroup

	// Start all children with their own restart loops
	for i := range states {
		waitgroup.Add(1)
		go func(idx int) {
			defer waitgroup.Done()
			sup.runChildLoop(ctx, states[idx])
		}(i)
	}

	waitgroup.Wait()
	return nil
}

// runChildLoop manages the lifecycle of a single child with restarts
func (sup Supervisor) runChildLoop(ctx context.Context, state *childState) {
	for {
		if ctx.Err() != nil {
			return
		}

		// Create child context
		childCtx, cancel := context.WithCancel(ctx)
		state.mu.Lock()
		state.cancel = cancel
		state.mu.Unlock()

		// Run the child with panic recovery
		err := safeRunWorker(childCtx, state.spec)
		cancel()

		if ctx.Err() != nil {
			return
		}

		// Log exit
		if err != nil {
			log.Printf("supervisor: child '%s' crashed: %v", state.spec.Name, err)
		} else {
			log.Printf("supervisor: child '%s' exited cleanly", state.spec.Name)
		}

		if !policyRestart(state.spec.Restart, err) {
			log.Printf("supervisor: child '%s' will not be restarted (policy)", state.spec.Name)
			return
		}

		if !recordTrackerRestart(&state.tracker, time.Now()) {
			log.Printf("supervisor: child '%s' exceeded max restarts (%d in %v), giving up",
				state.spec.Name, sup.Config.MaxRestarts, sup.Config.MaxTime)
			return
		}

		log.Printf("supervisor: restarting child '%s' after %v", state.spec.Name, sup.Config.RestartDelay)
		select {
		case <-time.After(sup.Config.RestartDelay):
		case <-ctx.Done():
			return
		}
	}
}

func safeRunWorker(ctx context.Context, child ChildSpec) error {
	var panicErr error

	defer func() {
		panicValue := recover()
		if panicValue != nil {
			panicErr = fmt.Errorf("worker '%s' panic: %v", child.Name, panicValue)
		}
	}()

	workerErr := child.Worker.Run(ctx)

	if panicErr != nil {
		return panicErr
	}
	return workerErr
}
