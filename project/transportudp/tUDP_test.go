package TransportUDP

import (
	"testing"
	"fmt"
	"time"
	// peermonitor "project/peermonitor" //will only be using the types in peermonitor...
	ordersync "project/ordersync" //...and ordersync
	// "reflect"
)

func TestMainLike(t *testing.T) {
	const peerID = 1

	// Channels for networking
	// eId := ordersync.ElevID(peerID)
	// fmt.Print("elevatorID: ", eId)
	// recTx := make(chan peermonitor.RecoveryMsg)
	// recRx := make(chan peermonitor.RecoveryMsg)

	osNetMsgTx := make(chan ordersync.NetMsg)
	osNetMsgRx := make(chan ordersync.NetMsg)

	pmNetMsgRx := make(chan ordersync.NetMsg)

	go Run(peerID, osNetMsgTx, osNetMsgRx, pmNetMsgRx)

	go func() {
		var netMsg = ordersync.NetMsg{}
		for {
			osNetMsgTx <- netMsg
			time.Sleep(1 * time.Second)
		}
	}()

	// go func() {
	// 	var recMsg = peermonitor.RecoveryMsg{}
	// 	for {
	// 		recTx <- recMsg
	// 		time.Sleep(1 * time.Second)
	// 	}
	// }()

	fmt.Println("Started")
	for {
		select {
		// case <-recRx:
		// 	fmt.Printf("Recovery message recieved\n")
		// 	fmt.Println()

		case <-osNetMsgRx:
			fmt.Printf("Net-msg recieved on order-sync chan\n")
			fmt.Println()

		case <-pmNetMsgRx:
			fmt.Printf("Net-msg recieved on peermonitor chan\n")
			fmt.Println()
		}
	}
}