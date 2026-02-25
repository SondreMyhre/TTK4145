package transportudp

import (
	"testing"
	"fmt"
	"time"
	peermonitor "project/peermonitor" //will only be using the types in peermonitor...
	ordersync "project/ordersync" //...and ordersync
	// "reflect"
)

func TestMainLike(t *testing.T) {
	OrderSyncTx := make(chan ordersync.NetMsg)
	OrderSyncRx := make(chan ordersync.NetMsg)
	PeerMonitorTx := make(chan peermonitor.HeartBeat)
	PeerMonitorRx := make(chan peermonitor.HeartBeat)

	go Run(OrderSyncTx, OrderSyncRx, PeerMonitorTx, PeerMonitorRx)

	go func() {
		var netMsg = ordersync.NetMsg{}
		for {
			OrderSyncTx <- netMsg
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		heartBeat := peermonitor.HeartBeat{SenderID: 1}
		for {
			PeerMonitorTx <- heartBeat
			time.Sleep(1 * time.Second)
		}
	}()

	fmt.Println("Started")
	for {
		select {
		case <-OrderSyncRx:
			fmt.Println("															Net-msg recieved on order-sync chan")
			fmt.Println()

		case heartbeat := <-PeerMonitorRx:
			fmt.Println("Heartbeat recieved on peermonitor chan")
			fmt.Printf("		SenderID: %v", heartbeat.SenderID)
			fmt.Println()
		}
	}
}