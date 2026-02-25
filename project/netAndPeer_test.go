package main

import (
	"fmt"
	transportudp "project/transportudp"
	ordersync "project/ordersync"     //...and ordersync
	peermonitor "project/peermonitor" //will only be using the types in peermonitor...
	"testing"
	"time"
	// "reflect"
)

func TestSystem(t *testing.T) {
	OrderSyncTx := make(chan ordersync.NetMsg)
	OrderSyncRx := make(chan ordersync.NetMsg)
	PeerMonitorTx := make(chan peermonitor.HeartBeat)
	PeerMonitorRx := make(chan peermonitor.HeartBeat)

	peerIDInt := 1

	go transportudp.Run(OrderSyncTx, OrderSyncRx, PeerMonitorTx, PeerMonitorRx)

	peermonitorConfig := peermonitor.PeerConfig{Timeout: 10 * time.Second, TickInterval: 50 * time.Millisecond}
	go peermonitor.Run(peerIDInt, peermonitorConfig, PeerMonitorRx, PeerMonitorTx, peerEventChan)
	// TO-DO: Fix peermonitor, so peermonitor and transportUDP can be tested together
}