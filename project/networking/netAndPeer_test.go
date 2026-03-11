package networking

import (
	"context"
	ordersync "project/ordersync"
	peermonitor "project/peermonitor"
	"testing"
	"time"
)

func TestSystem(t *testing.T) {
	OrderSyncTx := make(chan ordersync.NetMsg)
	OrderSyncRx := make(chan ordersync.NetMsg)
	PeerMonitorTx := make(chan peermonitor.HeartBeat)
	PeerMonitorRx := make(chan peermonitor.HeartBeat)
	peerEventChan := make(chan []ordersync.PeerUpdate)

	peerID := "1"

	peermonitorConfig := peermonitor.PeerConfig{Timeout: 10 * time.Second, TickInterval: 50 * time.Millisecond, HeartBeatTicker: 1 * time.Second}
	ctx := context.Background()

	go Run(ctx, OrderSyncTx, OrderSyncRx, PeerMonitorTx, PeerMonitorRx)
	go peermonitor.Run(peerID, ctx, peermonitorConfig, PeerMonitorRx, PeerMonitorTx, peerEventChan)
	select {}
	// TO-DO: Fix peermonitor, so peermonitor and networking can be tested together
}
