package transportudp
// REMEMBER:
// maps need to have string-keys to be broadcasted
// all structs to be sent must have public members
// broadcast-ip defined in bcast.go is different for personalPCs and labPC
	// broadcast-ip on labpc: 10.100.23.255

import (
	"project/transportudp/bcast"
	ordersync "project/ordersync" //...and ordersync
)

const broadcastPort = 50000

func Run(
		NetMsgTx <- chan ordersync.NetMsg, //broadcast netMsg from ordersync
		osNetMsgRx chan<- ordersync.NetMsg, //send NetMsg to ordersync ...
		pmNetMsgRx chan<- ordersync.NetMsg, //... and peermonitor
	) {
	netMsgRx := make(chan ordersync.NetMsg)

	// Reads messages from the channels, decodes them, and broadcasts
	go bcast.Transmitter(broadcastPort, NetMsgTx)

	// Reads messages from the network, decodes them and, send over respective channels
	go bcast.Receiver(broadcastPort, netMsgRx)

	// Broadcasted net-msgs should be directed to both ordersync and peermonitor
	mergeNetChans(netMsgRx, osNetMsgRx, pmNetMsgRx)
}

func mergeNetChans(netMsgCh <-chan ordersync.NetMsg, osNetMsgRx chan<- ordersync.NetMsg, pmNetMsgRx chan<- ordersync.NetMsg) {
	for{
		netmsg := <-netMsgCh
		pmNetMsgRx <- netmsg
		osNetMsgRx <- netmsg
	}
}
