package networking

// REMEMBER:
// maps need to have string-keys to be broadcasted
// all structs to be sent must have public members
// broadcast-ip defined in bcast.go is different for personalPCs and labPC
// broadcast-ip on labpc: 10.100.23.255

import (
	bcast "project/networking/bcast"
	ordersync "project/ordersync"
	peermonitor "project/peermonitor"
)

const broadcastPort = 50000

func Run(
	ordersyncTx <-chan ordersync.NetMsg,
	peermonitorTx <-chan peermonitor.HeartBeat,

	ordersyncRx chan<- ordersync.NetMsg,
	peermonitorRx chan<- peermonitor.HeartBeat,
) {
	go bcast.Transmitter(broadcastPort, ordersyncTx, peermonitorTx)
	go bcast.Receiver(broadcastPort, ordersyncRx, peermonitorRx)

	select {}
}
