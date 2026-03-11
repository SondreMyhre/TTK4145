package supervisor

import (
	"context"
	"time"
)

// WorkerFunc implements Worker interface
func (fn WorkerFunc) Run(ctx context.Context) error { return fn(ctx) }

// New creates a supervisor with default config
func New(children []ChildSpec) Supervisor {
	return Supervisor{
		Children: children,
		Config: SupervisorConfig{
			MaxRestarts:  DefaultMaxRestarts,
			MaxTime:      DefaultMaxTime,
			RestartDelay: DefaultRestartDelay,
		},
	}
}

// NewWithConfig creates a supervisor with custom config
func NewWithConfig(children []ChildSpec, config SupervisorConfig) Supervisor {
	return Supervisor{
		Children: children,
		Config:   config,
	}
}

// shouldRestart determines if a child should be restarted based on policy and error
func shouldRestart(policy RestartPolicy, err error) bool {
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
func newRestartTracker(maxCount int, window time.Duration) *restartTracker {
	return &restartTracker{
		timestamps: make([]time.Time, 0, maxCount),
		maxCount:   maxCount,
		window:     window,
	}
}

// recordRestart adds a restart timestamp and returns true if within limits
func (restartTracker *restartTracker) recordRestart(now time.Time) bool {
	cutoff := now.Add(-restartTracker.window)

	// Remove old timestamps outside the window
	valid := make([]time.Time, 0, len(restartTracker.timestamps))
	for _, timeStamps := range restartTracker.timestamps {
		if timeStamps.After(cutoff) {
			valid = append(valid, timeStamps)
		}
	}
	restartTracker.timestamps = valid

	// Check if we've exceeded max restarts
	if len(restartTracker.timestamps) >= restartTracker.maxCount {
		return false
	}

	restartTracker.timestamps = append(restartTracker.timestamps, now)
	return true
}
