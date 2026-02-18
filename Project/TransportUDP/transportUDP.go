package TransportUDP

import (
	"Project/TransportUDP/bcast"
	"Project/PeerMonitor" //will only be using the types in peermonitor and ordersync
	"Project/OrderSync"
	// "flag"
	// "fmt"
	// "os"
	// "time"
)

// TO-DO: Set up Msg-members as OrderSync and PeerMonitor needs
// NOTE: all members we want to broadcast has to be public!!!
type PeerMonitorMsg struct {
	Message string
	Iter int
}

type OrderSyncMsg struct {
	Message string
	Iter int
}

// Run is called in main.go
func Run(PeerMonitorTx <-chan PeerMonitorMsg, 
		OrderSyncTx <-chan OrderSyncMsg, 

		PeerMonitorRx chan<- PeerMonitorMsg,
		OrderSyncRx chan<- OrderSyncMsg,

		port int,
	) {
	// Reads messages from the channels, decodes them, and broadcasts
	go bcast.Transmitter(port, PeerMonitorTx, OrderSyncTx)

	// Reads messages from the network, decodes them and, send over respective channels
	go bcast.Receiver(port, PeerMonitorRx, OrderSyncRx)
}