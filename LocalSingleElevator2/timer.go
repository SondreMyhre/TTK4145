package localsingle2

import (
	"sync"
	"time"
)

// Timer provides a simple countdown timer with thread-safe access.
type Timer struct {
	mu      sync.Mutex
	endTime time.Time
	active  bool
}

// globalTimer is the package-level timer instance.
var globalTimer Timer

// TimerStart starts the timer for the given duration in seconds.
func TimerStart(durationSeconds float64) {
	globalTimer.mu.Lock()
	defer globalTimer.mu.Unlock()
	globalTimer.endTime = time.Now().Add(time.Duration(durationSeconds * float64(time.Second)))
	globalTimer.active = true
}

// TimerStop stops the timer.
func TimerStop() {
	globalTimer.mu.Lock()
	defer globalTimer.mu.Unlock()
	globalTimer.active = false
}

// TimerTimedOut returns true if the timer is active and has expired.
func TimerTimedOut() bool {
	globalTimer.mu.Lock()
	defer globalTimer.mu.Unlock()
	return globalTimer.active && time.Now().After(globalTimer.endTime)
}
