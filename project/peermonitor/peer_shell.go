package peermonitor

import (
	"time"
)

func Run(
	peerID string,

	heartBeatRx <-chan HeartBeat,

	ordersyncTx chan<- PeerMsg,
	heartBeatTx chan<- HeartBeat,
) error {
	peerTicker := time.NewTicker(PEER_TICK_INTERVAL)
	heartBeatTicker := time.NewTicker(HEART_BEAT_TICK_INTERVAL)
	defer peerTicker.Stop()
	defer heartBeatTicker.Stop()
	peerList := make([]Peer, 0)

	for {
		select {
		case msg := <-heartBeatRx:
			var changed bool
			now := time.Now()
			peerList, changed = HandleHeartbeats(peerList, msg, now)
			if changed {
				ordersyncTx <- ToPeerUpdate(peerList)
			}

		case <-peerTicker.C:
			var timeoutChanged bool
			now := time.Now()
			peerList, timeoutChanged = CheckTimeouts(peerList, now, PEER_TIMEOUT)
			if timeoutChanged {
				ordersyncTx <- ToPeerUpdate(peerList)
			}

		case <-heartBeatTicker.C:
			heartBeat := HeartBeat{SenderID: ElevID(peerID)}
			heartBeatTx <- heartBeat
		}
	}
}
