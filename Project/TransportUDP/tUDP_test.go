package TransportUDP

import (
	"testing"
	"fmt"
	// "time"
	// peermonitor "Project/PeerMonitor" //will only be using the types in peermonitor...
	// ordersync "Project/OrderSync" //...and ordersync
	// "reflect"
)

func TestMainLike(t *testing.T) {
	const peerID = 1

	// Channels for networking
	// eId := ordersync.ElevID(peerID)
	// fmt.Print("elevatorID: ", eId)
	recTx := make(chan peermonitor.RecoveryMsg)
	recRx := make(chan peermonitor.RecoveryMsg)

	osNetMsgTx := make(chan ordersync.NetMsg)
	osNetMsgRx := make(chan ordersync.NetMsg)

	pmNetMsgRx := make(chan ordersync.NetMsg)

	go Run(peerID, recTx, recRx, osNetMsgTx, osNetMsgRx, pmNetMsgRx)

	go func() {
		homatrix := HallOrderMatrix [N_FLOORS][N_HALL]orderMatrixEntry
		netMsg := ordersync.NetMsg{ordersync.ElevID(peerID), }
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