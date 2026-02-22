package TransportUDP
// REMEMBER:
// maps need to have string-keys to be broadcasted
// all structs to be sent must have public members
// 
// 

import (
	"project/transportudp/bcast"
	peermonitor "project/peermonitor" //will only be using the types in peermonitor...
	ordersync "project/ordersync" //...and ordersync
)

const portOffset = 60000

// Run is called in main.go
func Run(pID int,
		recTx <-chan peermonitor.RecoveryMsg, //broadcast recovery-msg from peermonitor
		recRx chan<- peermonitor.RecoveryMsg, //send recovery-msg to localsingle

		NetMsgTx <- chan ordersync.NetMsg, //broadcast netMsg from ordersync
		osNetMsgRx chan<- ordersync.NetMsg, //send NetMsg to ordersync ...
		pmNetMsgRx chan<- ordersync.NetMsg, //... and peermonitor
	) {
	// Declaring variables
	peerID := pID
	port := peerID + portOffset
	netMsgRx := make(chan<- ordersync.NetMsg)

	// Reads messages from the channels, decodes them, and broadcasts
	go bcast.Transmitter(port, recTx, NetMsgTx)

	// Reads messages from the network, decodes them and, send over respective channels
	// TO-DO: Recieved NetMsgs should be sent to both ordersync and peermonitor, make sure bcast.Recieve() supports this
	go bcast.Receiver(port, recRx, netMsgRx)

	// Broadcasted net-msgs should be directed to both ordersync and peermonitor
	go mergeNetChans(NetMsgTx, osNetMsgRx, pmNetMsgRx)
}

func mergeNetChans(netMsgCh <-chan ordersync.NetMsg, osNetMsgRx chan<- ordersync.NetMsg, pmNetMsgRx chan<- ordersync.NetMsg) {
	for{
		netmsg := <-netMsgCh
		pmNetMsgRx <- netmsg
		osNetMsgRx <- netmsg
	}
}
