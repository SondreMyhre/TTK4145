package supervisor

import (
	"context"
	"time"
)

// WorkerFunc implements Worker interface
func (fn WorkerFunc) Run(ctx context.Context) error { return fn(ctx) }

// NewSupervisor creates a supervisor with default config
func NewSupervisor(children []ChildSpec) Supervisor {
	return Supervisor{
		Children: children,
		Config: SupervisorConfig{
			MaxRestarts:  MAX_RESTARTS,
			MaxTime:      MAX_TIME,
			RestartDelay: RESTART_DELAY,
		},
	}
}

// shouldRestart determines if a child should be restarted based on policy and error
func policyRestart(policy RestartPolicy, err error) bool {
	switch policy {
	case Permanent:
		return true
	case Transient:
		return err != nil
	case Temporary:
		return false
	default:
		return err != nil
	}
}

// newRestartTracker creates a new restart rate limiter
func newRestartTracker(maxCount int, timewindow time.Duration) restartTracker {
	return restartTracker{
		crashCount:     0,
		firstCrashTime: time.Now(),
		maxCount:       maxCount,
		timewindow:     timewindow,
	}
}

// recordRestart increments the restart counter and returns true if within limits.
// It resets the counter if the time window has passed since the first crash.
func recordTrackerRestart(tracker *restartTracker, now time.Time) bool {
	// If the time window has passed, reset the counter
	if now.Sub(tracker.firstCrashTime) > tracker.timewindow {
		tracker.crashCount = 0
		tracker.firstCrashTime = now
	}

	tracker.crashCount++
	// Check if we've exceeded max restarts
	if tracker.crashCount > tracker.maxCount {
		return false
	}
	
	return true
}
