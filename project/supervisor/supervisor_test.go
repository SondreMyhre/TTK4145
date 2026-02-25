package supervisor

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	peermonitor "project/peermonitor"
	ordersync "project/ordersync"
	elevio "project/elevio"
	localsingle "project/localsingleelevator"
)

func TestSupervisor_RestartsPeerMonitorWhenHbRxClosed(t *testing.T) {
	hbRx := make(chan peermonitor.HeartBeat)
	close(hbRx)

	hbTx := make(chan peermonitor.HeartBeat, 10) // not used by test, but Run requires it
	chanOS := make(chan peermonitor.PeerMsg, 10)

	cfg := peermonitor.PeerConfig{
		Timeout:         50 * time.Millisecond,
		TickInterval:    10 * time.Millisecond,
		HeartBeatTicker: time.Hour, // prevent self-heartbeat goroutine from doing anything in test
	}

	var runs atomic.Int32
	w := WorkerFunc(func(ctx context.Context) error {
		runs.Add(1)
		return peermonitor.Run("self", ctx, cfg, hbRx, hbTx, chanOS)
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
		return peermonitor.Run("self", ctx, cfg, hbRx, hbTx, chanOS)
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
	hbRx := make(chan peermonitor.HeartBeat, 1)
	hbTx := make(chan peermonitor.HeartBeat, 10)

	// unbuffered output channel, and we never read from it
	chanOS := make(chan peermonitor.PeerMsg)

	cfg := peermonitor.PeerConfig{
		Timeout:         200 * time.Millisecond,
		TickInterval:    50 * time.Millisecond,
		HeartBeatTicker: time.Hour,
	}

	var runs atomic.Int32
	w := WorkerFunc(func(ctx context.Context) error {
		runs.Add(1)
		return peermonitor.Run("self", ctx, cfg, hbRx, hbTx, chanOS)
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

func TestSupervisor_RestartsOrderSyncWhenRxClosed(t *testing.T) {
	buttonChan := make(chan elevio.ButtonEvent)
	localStateChan := make(chan localsingle.ElevatorState)
	clearedOrdersChan := make(chan []localsingle.Order)

	rx := make(chan ordersync.NetMsg)
	close(rx)

	peerEventChan := make(chan []ordersync.Peer)
	localOrderChan := make(chan elevio.ButtonEvent)
	tx := make(chan ordersync.NetMsg, 10)
	lightCommandChan := make(chan elevio.DriverCommand, 10)

	var runs atomic.Int32
	w := WorkerFunc(func(ctx context.Context) error {
		runs.Add(1)
		return ordersync.Run(
			ctx,
			ordersync.ElevID("self"),
			buttonChan,
			localStateChan,
			clearedOrdersChan,
			rx,
			peerEventChan,
			localOrderChan,
			tx,
			lightCommandChan,
		)
	})

	sup := Supervisor{
		Child:        ChildSpec{Name: "ordersync", Worker: w, Restart: Transient},
		RestartDelay: 0,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
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
	peerEventChan := make(chan []ordersync.Peer)
	localOrderChan := make(chan elevio.ButtonEvent)
	tx := make(chan ordersync.NetMsg, 10)
	lightCommandChan := make(chan elevio.DriverCommand, 10)

	var runs atomic.Int32
	w := WorkerFunc(func(ctx context.Context) error {
		runs.Add(1)
		return ordersync.Run(
			ctx,
			ordersync.ElevID("self"),
			buttonChan,
			localStateChan,
			clearedOrdersChan,
			rx,
			peerEventChan,
			localOrderChan,
			tx,
			lightCommandChan,
		)
	})

	sup := Supervisor{
		Child:        ChildSpec{Name: "ordersync", Worker: w, Restart: Transient},
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

func TestSupervisor_OrderSyncStopsEvenIfTxSendWouldBlock(t *testing.T) {
	buttonChan := make(chan elevio.ButtonEvent, 1)
	localStateChan := make(chan localsingle.ElevatorState)
	clearedOrdersChan := make(chan []localsingle.Order)
	rx := make(chan ordersync.NetMsg)
	peerEventChan := make(chan []ordersync.Peer)
	localOrderChan := make(chan elevio.ButtonEvent)

	// Unbuffered tx channel + nobody reads => broadcast send would block
	tx := make(chan ordersync.NetMsg)
	lightCommandChan := make(chan elevio.DriverCommand, 10)

	var runs atomic.Int32
	w := WorkerFunc(func(ctx context.Context) error {
		runs.Add(1)
		return ordersync.Run(
			ctx,
			ordersync.ElevID("self"),
			buttonChan,
			localStateChan,
			clearedOrdersChan,
			rx,
			peerEventChan,
			localOrderChan,
			tx,
			lightCommandChan,
		)
	})

	sup := Supervisor{
		Child:        ChildSpec{Name: "ordersync", Worker: w, Restart: Transient},
		RestartDelay: 0,
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- sup.Run(ctx) }()

	// Hall button -> onHallButtonEvent -> broadcastNetMessage -> send on tx (blocks)
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