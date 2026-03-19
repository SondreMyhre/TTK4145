package peermonitor

import (
	"time"
	config "project/config"
)

func Run(
	peerID string,

	heartBeatRx <-chan HeartBeat,

	peerEventChan chan<- PeerMsg,
	heartBeatTx chan<- HeartBeat,
) {
	peerTicker := time.NewTicker(config.PEER_TICK_INTERVAL)
	heartBeatTicker := time.NewTicker(config.HEARTBEAT_TICK_INTERVAL)
	defer peerTicker.Stop()
	defer heartBeatTicker.Stop()
	peerList := make([]Peer, 0)

	for {
		select {
		case incomingHeartBeat := <-heartBeatRx:
			var isChanged bool
			now := time.Now()
			peerList, isChanged = HandleHeartbeats(peerList, incomingHeartBeat, now)
			if isChanged {
				peerEventChan <- ToPeerUpdate(peerList)
			}

		case <-peerTicker.C:
			var timeoutChanged bool
			now := time.Now()
			peerList, timeoutChanged = CheckTimeouts(peerList, now, config.PEER_TIMEOUT)
			if timeoutChanged {
				peerEventChan <- ToPeerUpdate(peerList)
			}

		case <-heartBeatTicker.C:
			outgoingHeartBeat := HeartBeat{SenderID: ElevID(peerID)}
			heartBeatTx <- outgoingHeartBeat
		}
	}
}
