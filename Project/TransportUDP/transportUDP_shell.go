package TransportUDP

import (
	"Project/TransportUDP/bcast"
	// "flag"
	// "fmt"
	// "os"
	// "time"
)

// find out what information NetMsg should carry
// has to contain all information to be sent over the network
// needs the orderMatrix from orderSync and a peer-message from peerMonitor
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