package peermonitor

import (
	"context"
	"testing"
	"time"
)

func TestRun_HeartbeatDoesNotSpamUpdates(t *testing.T) {
	const peerTick = 50 * time.Millisecond

	cfg := PeerConfig{
		Timeout:      10 * time.Second, // large so timeout can't happen during test
		TickInterval: peerTick,
		// Make self-heartbeats effectively "off" for the duration of the test.
		HeartBeatTicker: time.Hour,
	}

	hbRx := make(chan HeartBeat, 10)
	hbTx := make(chan HeartBeat, 10)
	chanOS := make(chan PeerMsg, 10)

	errCh := make(chan error, 1)
	done := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		errCh <- Run("self", ctx, cfg, hbRx, hbTx, chanOS)
		close(done)
	}()

	// First heartbeat -> expect exactly one update (new peer)
	hbRx <- HeartBeat{SenderID: "1"}

	select {
	case <-chanOS:
		// ok
	case <-time.After(300 * time.Millisecond):
		t.Fatalf("expected an update after first heartbeat")
	}

	// Drain any extra updates that may already be queued
Drain:
	for {
		select {
		case <-chanOS:
		default:
			break Drain
		}
	}

	// Second heartbeat soon after -> should not produce an update
	hbRx <- HeartBeat{SenderID: "1"}

	select {
	case <-chanOS:
		t.Fatalf("did not expect update on heartbeat that only refreshes LastSeen")
	case <-time.After(200 * time.Millisecond):
		// ok
	}

	close(hbRx)
	select {
	case <-done:
		// Run should return when hbRx is closed.
		// Current implementation returns a non-nil error in this case.
		_ = <-errCh
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("expected Run to return after hbRx is closed")
	}
}

func TestRun_TimeoutProducesUpdate(t *testing.T) {
	const peerTick = 50 * time.Millisecond

	cfg := PeerConfig{
		Timeout:      500 * time.Millisecond,
		TickInterval: peerTick,
		// Make self-heartbeats effectively "off" for the duration of the test.
		HeartBeatTicker: time.Hour,
	}

	hbRx := make(chan HeartBeat, 10)
	hbTx := make(chan HeartBeat, 10)
	chanOS := make(chan PeerMsg, 10)

	errCh := make(chan error, 1)
	done := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		errCh <- Run("self", ctx, cfg, hbRx, hbTx, chanOS)
		close(done)
	}()

	// Create peer
	hbRx <- HeartBeat{SenderID: "1"}

	// Consume the "new peer" update
	select {
	case <-chanOS:
	case <-time.After(300 * time.Millisecond):
		t.Fatalf("expected an update after first heartbeat")
	}

	// Now don't send anything. Wait long enough for timeout + ticker ticks.
	select {
	case upd := <-chanOS:
		found := false
		for _, p := range upd {
			if p.ID == "1" {
				found = true
				if p.PeerStatus != StatusDead {
					t.Fatalf("expected peer 1 Dead after timeout, got %v", p.PeerStatus)
				}
			}
		}
		if !found {
			t.Fatalf("expected to find peer 1 in update after timeout")
		}

	case <-time.After(900 * time.Millisecond):
		t.Fatalf("expected an update after timeout (no heartbeat sent)")
	}

	close(hbRx)
	select {
	case <-done:
		_ = <-errCh
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("expected Run to return after hbRx is closed")
	}
}
