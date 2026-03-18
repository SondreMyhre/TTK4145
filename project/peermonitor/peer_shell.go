package peermonitor

import (
	"context"
	"time"
)

func Run(ctx context.Context,
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
		case <-ctx.Done():
			return nil

		case msg := <-heartBeatRx:
			var changed bool
			now := time.Now()
			peerList, changed = HandleHeartbeats(peerList, msg, now)
			if changed {
				select {
				case ordersyncTx <- ToPeerUpdate(peerList):
				case <-ctx.Done():
					return nil
				}
			}

		case <-peerTicker.C:
			var timeoutChanged bool
			now := time.Now()
			peerList, timeoutChanged = CheckTimeouts(peerList, now, PEER_TIMEOUT)
			if timeoutChanged {
				select {
				case ordersyncTx <- ToPeerUpdate(peerList):
				case <-ctx.Done():
					return nil
				}
			}

		case <-heartBeatTicker.C:
			heartBeat := HeartBeat{SenderID: ElevID(peerID)}
			heartBeatTx <- heartBeat
		}
	}
}
