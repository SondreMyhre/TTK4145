package networking

import (
	"fmt"
	config "project/config"
	bcast "project/networking/bcast"
	ordersync "project/ordersync"
	peermonitor "project/peermonitor"
)

func Run(
	ordersyncTx <-chan ordersync.NetMsg,
	peermonitorTx <-chan peermonitor.HeartBeat,

	ordersyncRx chan<- ordersync.NetMsg,
	peermonitorRx chan<- peermonitor.HeartBeat,
) {
	broadcastSocket := fmt.Sprintf("%s:%d", config.BROADCAST_ADDRESS, config.BROADCAST_PORT)
	go bcast.Transmitter(broadcastSocket, config.BROADCAST_PORT, ordersyncTx, peermonitorTx)
	go bcast.Receiver(config.BROADCAST_PORT, ordersyncRx, peermonitorRx)

	select {}
}
