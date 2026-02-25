package transportudp

// REMEMBER:
// maps need to have string-keys to be broadcasted
// all structs to be sent must have public members
// broadcast-ip defined in bcast.go is different for personalPCs and labPC
// broadcast-ip on labpc: 10.100.23.255

import (
	ordersync "project/ordersync" //...and ordersync
	peermonitor "project/peermonitor"
	"project/transportudp/bcast"
)

const broadcastPort = 50000

func Run(
		ordersyncTx <- chan ordersync.NetMsg, //broadcast netMsg from ordersync
		ordersyncRx chan<- ordersync.NetMsg, //send NetMsg to ordersync ...

		peermonitorTx <- chan peermonitor.HeartBeat,
		peermonitorRx chan<- peermonitor.HeartBeat, //... and peermonitor
	) {
	// netMsgRx := make(chan ordersync.NetMsg)

	// Reads messages from the channels, decodes them, and broadcasts
	go bcast.Transmitter(broadcastPort, ordersyncTx, peermonitorTx)

	// Reads messages from the network, decodes them and, send over respective channels
	go bcast.Receiver(broadcastPort, ordersyncRx, peermonitorRx)
}
