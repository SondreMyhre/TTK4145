package TransportUDP

import (
	"testing"
	"fmt"
	"time"
	// "reflect"
)

func TestMainLike(t *testing.T) {
	const port = 60000

	// Channels for networking
	PeerMonitorTx := make(chan PeerMonitorMsg)
	PeerMonitorRx := make(chan PeerMonitorMsg)

	OrderSyncTx := make(chan OrderSyncMsg)
	OrderSyncRx := make(chan OrderSyncMsg)

	go Run(PeerMonitorTx, OrderSyncTx, PeerMonitorRx, OrderSyncRx, port)

	go func() {
		helloMsg := OrderSyncMsg{"Hello from ordersync", 0}
		for {
			helloMsg.Iter++
			OrderSyncTx <- helloMsg
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		helloMsg := PeerMonitorMsg{"Hello from ordersync", 0}
		for {
			helloMsg.Iter++
			PeerMonitorTx <- helloMsg
			time.Sleep(1 * time.Second)
		}
	}()

	fmt.Println("Started")
	for {
		select {
		case o := <-OrderSyncRx:
			fmt.Printf("Ordersync message:\n")
			fmt.Printf("  msg:            %v\n", o.Message)
			fmt.Printf("  iteration:      %v\n", o.Iter)
			fmt.Println()

		case p := <-PeerMonitorRx:
			fmt.Printf("PeerMonitor message:\n")
			fmt.Printf("  msg:            %v\n", p.Message)
			fmt.Printf("  iteration:      %v\n", p.Iter)
			fmt.Println()
		}
	}

	// select{}

	// t.Log("transportUDP-test ran.")
}