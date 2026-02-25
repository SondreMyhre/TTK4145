package peermonitor

import (
	"context"
	"fmt"
	"time"
)

// Shell PeerMonitor

func Run(peerID string, ctx context.Context, cfg PeerConfig, heartBeatRx <-chan HeartBeat, heartBeatTx chan<- HeartBeat, chanOS chan<- PeerMsg) error {
	ticker := time.NewTicker(cfg.TickInterval)
	defer ticker.Stop() //runs ticker while function is running

	peerList := make([]Peer, 0)

	// infinitely send heartBeats
	go sendHeartbeats(peerID, heartBeatTx)

	for {
		select {
		case <-ctx.Done():
			return nil

		case msg, ok := <-heartBeatRx:
			if !ok { // se om kanal er åpen, stopper dersom kanal er lukket
				return fmt.Errorf("peermonitor: heartBeatRx closed")
			}

			var changed bool
			now := time.Now()
			peerList, changed = HandleHeartbeats(peerList, msg, now) //looks for updates and sets changed to true/false
			if changed {
				//dont block forever in case noone is reading or buffer full
				select {
				case chanOS <- ToPeerUpdate(peerList):
				case <-ctx.Done():
					return nil
				}
			}

		case <-ticker.C: //C is channel for ticker
			// Periodically check for peers that have timed out (Alive -> Dead)
			var timeoutChanged bool
			now := time.Now()
			peerList, timeoutChanged = CheckTimeouts(peerList, now, cfg.Timeout) //looks for updates and sets changed to true/false
			if timeoutChanged {
				select {
				case chanOS <- ToPeerUpdate(peerList):
				case <-ctx.Done():
					return nil
				}
			}
		}
	}
}
