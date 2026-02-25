package supervisor

import (
	"context"
	"project/peermonitor"
	"sync/atomic"
	"testing"
	"time"
)

func TestSupervisor_RestartsPeerMonitorWhenHbRxClosed(t *testing.T) {
	hbRx := make(chan peermonitor.NetMsg)
	close(hbRx)

	chanOS := make(chan peermonitor.PeerMsg, 10)

	cfg := peermonitor.PeerConfig{
		Timeout:      50 * time.Millisecond,
		TickInterval: 10 * time.Millisecond,
	}

	var runs atomic.Int32
	w := WorkerFunc(func(ctx context.Context) error {
		runs.Add(1)
		return peermonitor.Run(ctx, cfg, hbRx, chanOS)
	})

	sup := Supervisor{
		Child: ChildSpec{
			Name:    "peermonitor",
			Worker:  w,
			Restart: Transient,
		},
		RestartDelay: 0,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()

	_ = sup.Run(ctx)

	if got := runs.Load(); got < 2 {
		t.Fatalf("expected peermonitor to be restarted (>=2 runs), got %d", got)
	}
}

func TestSupervisor_StopsCleanlyOnContextCancel(t *testing.T) {
	hbRx := make(chan peermonitor.NetMsg)
	defer close(hbRx)

	chanOS := make(chan peermonitor.PeerMsg, 10)

	cfg := peermonitor.PeerConfig{
		Timeout:      50 * time.Millisecond,
		TickInterval: 10 * time.Millisecond,
	}

	var runs atomic.Int32
	w := WorkerFunc(func(ctx context.Context) error {
		runs.Add(1)
		return peermonitor.Run(ctx, cfg, hbRx, chanOS)
	})

	sup := Supervisor{
		Child: ChildSpec{
			Name:    "peermonitor",
			Worker:  w,
			Restart: Transient,
		},
		RestartDelay: 0,
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)

	go func() { errCh <- sup.Run(ctx) }()

	time.Sleep(10 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("expected nil on cancel shutdown, got %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("supervisor did not stop after context cancel")
	}

	if got := runs.Load(); got != 1 {
		t.Fatalf("expected exactly 1 run (no restart), got %d", got)
	}
}

func TestSupervisor_PeerMonitorStopsEvenIfOutputSendWouldBlock(t *testing.T) {
	hbRx := make(chan peermonitor.NetMsg, 1)

	// unbuffered output channel, and we never read from it
	chanOS := make(chan peermonitor.PeerMsg)

	cfg := peermonitor.PeerConfig{
		Timeout:      200 * time.Millisecond,
		TickInterval: 50 * time.Millisecond,
	}

	var runs atomic.Int32
	w := WorkerFunc(func(ctx context.Context) error {
		runs.Add(1)
		return peermonitor.Run(ctx, cfg, hbRx, chanOS)
	})

	sup := Supervisor{
		Child: ChildSpec{
			Name:    "peermonitor",
			Worker:  w,
			Restart: Transient,
		},
		RestartDelay: 0,
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- sup.Run(ctx) }()

	// Trigger changed == true => tries to send update to chanOS (but nobody reads)
	hbRx <- peermonitor.NetMsg{SenderID: peermonitor.ElevID("peer-1")}

	time.Sleep(10 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("expected nil on cancel shutdown, got %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("supervisor did not stop after cancel while output send would block")
	}

	if got := runs.Load(); got != 1 {
		t.Fatalf("expected exactly 1 run (no restart), got %d", got)
	}
}

func TestSupervisor_RestartsOnPanic(t *testing.T) {
	var runs atomic.Int32

	w := WorkerFunc(func(ctx context.Context) error {
		runs.Add(1)
		panic("boom")
	})

	sup := Supervisor{
		Child:        ChildSpec{Name: "panic", Worker: w, Restart: Transient},
		RestartDelay: 0,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_ = sup.Run(ctx)

	if runs.Load() < 2 {
		t.Fatalf("expected restart after panic (>=2 runs), got %d", runs.Load())
	}
}
