package supervisor

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	elevio "project/elevio"
	localsingle "project/localsingle"
	networking "project/networking"
	ordersync "project/ordersync"
	peermonitor "project/peermonitor"
)

// Helper to create single-child supervisor for tests
func newTestSupervisor(child ChildSpec) Supervisor {
	return Supervisor{
		Children: []ChildSpec{child},
		Config: SupervisorConfig{
			MaxRestarts:  10,
			MaxTime:      time.Minute,
			RestartDelay: 0,
		},
	}
}

func TestSupervisor_RestartsPeerMonitorWhenHbRxClosed(t *testing.T) {
	hbRx := make(chan peermonitor.HeartBeat)
	close(hbRx)

	hbTx := make(chan peermonitor.HeartBeat, 10)
	chanOS := make(chan peermonitor.PeerMsg, 10)

	cfg := peermonitor.PeerConfig{
		Timeout:         50 * time.Millisecond,
		TickInterval:    10 * time.Millisecond,
		HeartBeatTicker: time.Hour,
	}

	var runs atomic.Int32
	w := WorkerFunc(func(ctx context.Context) error {
		runs.Add(1)
		return peermonitor.Run(ctx, "self", cfg, hbRx, hbTx, chanOS)
	})

	sup := newTestSupervisor(ChildSpec{
		Name:    "peermonitor",
		Worker:  w,
		Restart: Transient,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_ = sup.Run(ctx)

	if got := runs.Load(); got < 2 {
		t.Fatalf("expected peermonitor to be restarted (>=2 runs), got %d", got)
	}
}

func TestSupervisor_StopsCleanlyOnContextCancel(t *testing.T) {
	hbRx := make(chan peermonitor.HeartBeat)
	defer close(hbRx)

	hbTx := make(chan peermonitor.HeartBeat, 10)
	chanOS := make(chan peermonitor.PeerMsg, 10)

	cfg := peermonitor.PeerConfig{
		Timeout:         50 * time.Millisecond,
		TickInterval:    10 * time.Millisecond,
		HeartBeatTicker: time.Hour,
	}

	var runs atomic.Int32
	w := WorkerFunc(func(ctx context.Context) error {
		runs.Add(1)
		return peermonitor.Run(ctx, "self", cfg, hbRx, hbTx, chanOS)
	})

	sup := newTestSupervisor(ChildSpec{
		Name:    "peermonitor",
		Worker:  w,
		Restart: Transient,
	})

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
	hbRx := make(chan peermonitor.HeartBeat, 1)
	hbTx := make(chan peermonitor.HeartBeat, 10)
	chanOS := make(chan peermonitor.PeerMsg)

	cfg := peermonitor.PeerConfig{
		Timeout:         200 * time.Millisecond,
		TickInterval:    50 * time.Millisecond,
		HeartBeatTicker: time.Hour,
	}

	var runs atomic.Int32
	w := WorkerFunc(func(ctx context.Context) error {
		runs.Add(1)
		return peermonitor.Run(ctx, "self", cfg, hbRx, hbTx, chanOS)
	})

	sup := newTestSupervisor(ChildSpec{
		Name:    "peermonitor",
		Worker:  w,
		Restart: Transient,
	})

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- sup.Run(ctx) }()

	hbRx <- peermonitor.HeartBeat{SenderID: peermonitor.ElevID("peer-1")}

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

	sup := newTestSupervisor(ChildSpec{
		Name:    "panic",
		Worker:  w,
		Restart: Transient,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_ = sup.Run(ctx)

	if runs.Load() < 2 {
		t.Fatalf("expected restart after panic (>=2 runs), got %d", runs.Load())
	}
}

func TestSupervisor_RestartsOrderSyncWhenRxClosed(t *testing.T) {
	buttonChan := make(chan elevio.ButtonEvent)
	localStateChan := make(chan localsingle.ElevatorState)
	clearedOrdersChan := make(chan []localsingle.Order)

	rx := make(chan ordersync.NetMsg)
	close(rx)

	peerEventChan := make(chan []ordersync.PeerUpdate)
	tx := make(chan ordersync.NetMsg, 10)
	lightCommandChan := make(chan elevio.DriverCommand, 10)
	worldviewChan := make(chan ordersync.WorldviewMsg, 1)

	var runs atomic.Int32
	w := WorkerFunc(func(ctx context.Context) error {
		runs.Add(1)
		return ordersync.RunWorldView(
			ctx,
			ordersync.ElevID("self"),
			buttonChan,
			localStateChan,
			clearedOrdersChan,
			rx,
			peerEventChan,
			tx,
			lightCommandChan,
			worldviewChan,
		)
	})

	sup := newTestSupervisor(ChildSpec{
		Name:    "ordersync",
		Worker:  w,
		Restart: Transient,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_ = sup.Run(ctx)

	if got := runs.Load(); got < 2 {
		t.Fatalf("expected ordersync to be restarted (>=2 runs), got %d", got)
	}
}

func TestSupervisor_OrderSyncStopsCleanlyOnContextCancel(t *testing.T) {
	buttonChan := make(chan elevio.ButtonEvent)
	localStateChan := make(chan localsingle.ElevatorState)
	clearedOrdersChan := make(chan []localsingle.Order)
	rx := make(chan ordersync.NetMsg)
	peerEventChan := make(chan []ordersync.PeerUpdate)
	tx := make(chan ordersync.NetMsg, 10)
	lightCommandChan := make(chan elevio.DriverCommand, 10)
	worldviewChan := make(chan ordersync.WorldviewMsg, 1)

	var runs atomic.Int32
	w := WorkerFunc(func(ctx context.Context) error {
		runs.Add(1)
		return ordersync.RunWorldView(
			ctx,
			ordersync.ElevID("self"),
			buttonChan,
			localStateChan,
			clearedOrdersChan,
			rx,
			peerEventChan,
			tx,
			lightCommandChan,
			worldviewChan,
		)
	})

	sup := newTestSupervisor(ChildSpec{
		Name:    "ordersync",
		Worker:  w,
		Restart: Transient,
	})

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

func TestSupervisor_OrderSyncStopsEvenIfTxSendWouldBlock(t *testing.T) {
	buttonChan := make(chan elevio.ButtonEvent, 1)
	localStateChan := make(chan localsingle.ElevatorState)
	clearedOrdersChan := make(chan []localsingle.Order)
	rx := make(chan ordersync.NetMsg)
	peerEventChan := make(chan []ordersync.PeerUpdate)
	tx := make(chan ordersync.NetMsg)
	lightCommandChan := make(chan elevio.DriverCommand, 10)
	worldviewChan := make(chan ordersync.WorldviewMsg)

	var runs atomic.Int32
	w := WorkerFunc(func(ctx context.Context) error {
		runs.Add(1)
		return ordersync.RunWorldView(
			ctx,
			ordersync.ElevID("self"),
			buttonChan,
			localStateChan,
			clearedOrdersChan,
			rx,
			peerEventChan,
			tx,
			lightCommandChan,
			worldviewChan,
		)
	})

	sup := newTestSupervisor(ChildSpec{
		Name:    "ordersync",
		Worker:  w,
		Restart: Transient,
	})

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- sup.Run(ctx) }()

	buttonChan <- elevio.ButtonEvent{Floor: 0, Button: elevio.BT_HallUp}

	time.Sleep(10 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("expected nil on cancel shutdown, got %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("supervisor did not stop after cancel while tx send would block")
	}

	if got := runs.Load(); got != 1 {
		t.Fatalf("expected exactly 1 run (no restart), got %d", got)
	}
}

func TestSupervisor_MultipleChildren(t *testing.T) {
	var runs1, runs2 atomic.Int32

	sup := Supervisor{
		Children: []ChildSpec{
			{
				Name: "child1",
				Worker: WorkerFunc(func(ctx context.Context) error {
					runs1.Add(1)
					<-ctx.Done()
					return nil
				}),
				Restart: Permanent,
			},
			{
				Name: "child2",
				Worker: WorkerFunc(func(ctx context.Context) error {
					runs2.Add(1)
					<-ctx.Done()
					return nil
				}),
				Restart: Permanent,
			},
		},
		Config: SupervisorConfig{
			MaxRestarts:  5,
			MaxTime:      time.Minute,
			RestartDelay: 0,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_ = sup.Run(ctx)

	if runs1.Load() < 1 || runs2.Load() < 1 {
		t.Fatalf("expected both children to run, got runs1=%d, runs2=%d", runs1.Load(), runs2.Load())
	}
}

func TestSupervisor_MaxRestartsExceeded(t *testing.T) {
	var runs atomic.Int32

	sup := Supervisor{
		Children: []ChildSpec{
			{
				Name: "crasher",
				Worker: WorkerFunc(func(ctx context.Context) error {
					runs.Add(1)
					return nil // Exits immediately (error = nil, but Permanent means restart)
				}),
				Restart: Permanent,
			},
		},
		Config: SupervisorConfig{
			MaxRestarts:  3,
			MaxTime:      time.Second,
			RestartDelay: 0,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = sup.Run(ctx)

	// Should stop after MaxRestarts+1 (initial + 3 restarts = 4)
	if got := runs.Load(); got > 4 {
		t.Fatalf("expected max ~4 runs due to rate limiting, got %d", got)
	}
}

// Dummy use of networking to avoid unused import error
var _ = networking.Run
