package networking

import (
	"fmt"
	config "project/config"
	bcast "project/networking/bcast"
	ordersync "project/ordersync"
	peermonitor "project/peermonitor"
)

func Run(
	worldViewTx <-chan ordersync.NetMsg,
	peermonitorTx <-chan peermonitor.HeartBeat,

	worldViewRx chan<- ordersync.NetMsg,
	peermonitorRx chan<- peermonitor.HeartBeat,
) {
	broadcastSocket := fmt.Sprintf("%s:%d", config.BROADCAST_ADDRESS, config.BROADCAST_PORT)
	go bcast.Transmitter(broadcastSocket, config.BROADCAST_PORT, worldViewTx, peermonitorTx)
	go bcast.Receiver(config.BROADCAST_PORT, worldViewRx, peermonitorRx)

	select {}
}
