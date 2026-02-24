package peermonitor

import (
	ordersync "project/ordersync" // ⚠️ bytt hvis go.mod har annet module-navn
	"testing"
	"time"
)


func TestRun_HeartbeatDoesNotSpamUpdates(t *testing.T) {
	// Huge timeout so ticker cannot mark peer dead during this test
	const peerTick = 50 * time.Millisecond

<<<<<<< HEAD:Project/PeerMonitor/peer_test.go
	cfg := PeerConfig{
    	Timeout:      500 * time.Millisecond,
    	TickInterval: peerTick,
}
	hbRx := make(chan shared.NetMsg, 10)
	chanOS := make(chan PeerUpdate, 10)
=======
	hbRx := make(chan ordersync.NetMsg, 10)
	chanOS := make(chan []ordersync.Peer, 10)
>>>>>>> c0ac4cfc61b971fec8d12f6737589e3e49274151:project/peermonitor/peer_test.go

	done := make(chan struct{})
	go func() {
		Run(cfg, hbRx, chanOS)
		close(done)
	}()

	// First heartbeat -> expect exactly one update (new peer)
	hbRx <- ordersync.NetMsg{SenderID: "1"}

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
	hbRx <- ordersync.NetMsg{SenderID: "1"}

	select {
	case <-chanOS:
		t.Fatalf("did not expect update on heartbeat that only refreshes LastSeen")
	case <-time.After(200 * time.Millisecond):
		// ok
	}

	close(hbRx)
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("expected Run to return after hbRx is closed")
	}
}

func TestRun_TimeoutProducesUpdate(t *testing.T) {
	// Small timeout so peer becomes dead quickly
	const peerTick = 50 * time.Millisecond

	cfg := PeerConfig{
    	Timeout:      500 * time.Millisecond,
    	TickInterval: peerTick,
}

	hbRx := make(chan ordersync.NetMsg, 10)
	chanOS := make(chan []ordersync.Peer, 10)

	done := make(chan struct{})
	go func() {
		Run(cfg, hbRx, chanOS)
		close(done)
	}()

	// Create peer
	hbRx <- ordersync.NetMsg{SenderID: "1"}

	// Consume the "new peer" update
	select {
	case <-chanOS:
	case <-time.After(300 * time.Millisecond):
		t.Fatalf("expected an update after first heartbeat")
	}

	// Now don't send anything. Wait long enough for timeout + a couple ticker ticks.
	// (Ticker is 50ms in Run, so 400ms is plenty even on Windows.)
	select {
	case upd := <-chanOS:
		// Expect peer 1 to be Dead in some update after timeout
		found := false
		for _, p := range upd {
			if p.ID == "1" {
				found = true
<<<<<<< HEAD:Project/PeerMonitor/peer_test.go
				if p.PeerStatus != StatusDead {
					t.Fatalf("expected peer 1 Dead after timeout, got %v", p.PeerStatus)
=======
				if p.Status != ordersync.Dead {
					t.Fatalf("expected peer 1 Dead after timeout, got %v", p.Status)
>>>>>>> c0ac4cfc61b971fec8d12f6737589e3e49274151:project/peermonitor/peer_test.go
				}
			}
		}
		if !found {
			t.Fatalf("expected peer 1 to exist in timeout update")
		}
	case <-time.After(600 * time.Millisecond):
		t.Fatalf("expected an update after timeout (no heartbeat sent)")
	}

	close(hbRx)
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("expected Run to return after hbRx is closed")
	}
}
