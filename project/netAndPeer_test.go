package main

import (
	"context"
	"testing"
	"time"

	networking "project/networking"
	ordersync "project/ordersync"
	peermonitor "project/peermonitor"
)

func TestSystem(t *testing.T) {
	OrderSyncTx := make(chan ordersync.NetMsg)
	OrderSyncRx := make(chan ordersync.NetMsg)
	PeerMonitorTx := make(chan peermonitor.HeartBeat)
	PeerMonitorRx := make(chan peermonitor.HeartBeat)
	peerEventChan := make(chan peermonitor.PeerMsg, 10) // IMPORTANT: buffer helps
	peerID := "1"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go networking.Run(ctx, OrderSyncTx, OrderSyncRx, PeerMonitorTx, PeerMonitorRx)

	cfg := peermonitor.PeerConfig{
		Timeout:      10 * time.Second,
		TickInterval: 50 * time.Millisecond,
		HeartBeatTicker: time.Hour,
		// make sure HeartBeatTicker is set too (see next section)
	}

	go func() {
		_ = peermonitor.Run(ctx, peerID, cfg, PeerMonitorRx, PeerMonitorTx, peerEventChan)
	}()

	// Actually observe something from peerEventChan (otherwise this isn't a test)
	select {
	case msg := <-peerEventChan:
		_ = msg // assert something meaningful here
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for peer update")
	}
}