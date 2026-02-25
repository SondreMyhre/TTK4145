package transportudp

import (
	"testing"
	"fmt"
	"time"
	// peermonitor "project/peermonitor" //will only be using the types in peermonitor...
	ordersync "project/ordersync" //...and ordersync
	// "reflect"
)

func TestMainLike(t *testing.T) {
	osNetMsgTx := make(chan ordersync.NetMsg)
	osNetMsgRx := make(chan ordersync.NetMsg)

	pmNetMsgRx := make(chan ordersync.NetMsg)

	go Run(osNetMsgTx, osNetMsgRx, pmNetMsgRx)

	go func() {
		var netMsg = ordersync.NetMsg{}
		for {
			osNetMsgTx <- netMsg
			time.Sleep(1 * time.Second)
		}
	}()

	fmt.Println("Started")
	for {
		select {
		case <-osNetMsgRx:
			fmt.Printf("Net-msg recieved on order-sync chan\n")
			fmt.Println()

		case <-pmNetMsgRx:
			fmt.Printf("Net-msg recieved on peermonitor chan\n")
			fmt.Println()
		}
	}
}