package peermonitor

import (
	"context"
	"fmt"
	"time"
)

func Run(cfg PeerConfig,
		peerID string, 
		ctx context.Context,  

		heartBeatRx <-chan HeartBeat, 

		ordersyncTx chan<- PeerMsg,
		heartBeatTx chan<- HeartBeat,
) error {
	peerTicker := time.NewTicker(cfg.TickInterval)
	heartBeatTicker := time.NewTicker(cfg.HeartBeatTicker)
	defer peerTicker.Stop()
	defer heartBeatTicker.Stop()
	peerList := make([]Peer, 0)

	for {
		select {
		case <-ctx.Done():
			return nil

		case msg, ok := <-heartBeatRx:
			if !ok {
				return fmt.Errorf("peermonitor: heartBeatRx closed")
			}

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
			peerList, timeoutChanged = CheckTimeouts(peerList, now, cfg.Timeout)
			if timeoutChanged {
				select {
				case ordersyncTx <- ToPeerUpdate(peerList):
				case <-ctx.Done():
					return nil
				}
			}

		case <- heartBeatTicker.C:
			heartBeat := HeartBeat{SenderID: ElevID(peerID)}
			heartBeatTx <- heartBeat
		}
	}
}
