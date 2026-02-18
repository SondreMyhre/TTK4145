package TransportUDP

import (
	"Project/TransportUDP/bcast"
	peermonitor "Project/PeerMonitor" //will only be using the types in peermonitor...
	ordersync "Project/OrderSync" //...and ordersync
)

const portOffset = 60000

// Run is called in main.go
func Run(pID int,
		recTx <-chan peermonitor.RecoveryMsg, //broadcast recovery-msg from peermonitor
		recRx chan<- peermonitor.RecoveryMsg, //send recovery-msg to localsingle

		osNetMsgTx <- chan ordersync.NetMsg, //broadcast netMsg from ordersync
		osNetMsgRx chan<- ordersync.NetMsg, //send NetMsg to ordersync ...

		pmNetMsgRx chan<- ordersync.NetMsg, //... and peermonitor
	) {
	// Declaring variables
	peerID := pID
	port := peerID + portOffset

	// Reads messages from the channels, decodes them, and broadcasts
	go bcast.Transmitter(port, recTx, osNetMsgTx)

	// Reads messages from the network, decodes them and, send over respective channels
	// TO-DO: Recieved NetMsgs should be sent to both ordersync and peermonitor, make sure bcast.Recieve() supports this
	go bcast.Receiver(port, recRx, osNetMsgRx, pmNetMsgRx)
}
