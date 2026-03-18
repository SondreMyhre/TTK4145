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
	config "project/config"
)

func Run(
	ordersyncTx <-chan ordersync.NetMsg,
	peermonitorTx <-chan peermonitor.HeartBeat,

	ordersyncRx chan<- ordersync.NetMsg,
	peermonitorRx chan<- peermonitor.HeartBeat,
) {
	go bcast.Transmitter(config.BROADCAST_PORT, ordersyncTx, peermonitorTx)
	go bcast.Receiver(config.BROADCAST_PORT, ordersyncRx, peermonitorRx)

	select {}
}
