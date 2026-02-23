package peermonitor

import (
	"time"
)

// Shell PeerMonitor

func Run(cfg PeerConfig, hbRx <-chan NetMsg, chanOS chan<- PeerUpdate) {
	var peerList []Peer

	ticker := time.NewTicker(cfg.TickInterval)//creates ticker* struct, ticker.C is channel
	defer ticker.Stop()   //runs ticker while function is running

	for {
		select {
		case msg, ok := <-hbRx:
			if !ok { // se om kanal er åpen, stopper dersom kanal er lukket
				return
			}
			var changed bool
			now := time.Now()
			peerList, changed = HandleHeartbeats(peerList, msg, now) //looks for updates and sets changed to true/false
			if changed {
				chanOS <- ToPeerUpdate(peerList)
			}
		case <-ticker.C: //C is channel for ticker 
		// Periodically check for peers that have timed out (Alive -> Dead)
			var timeoutChanged bool
			now := time.Now()
			peerList, timeoutChanged = CheckTimeouts(peerList, now, cfg.Timeout) //looks for updates and sets changed to true/false
			if timeoutChanged {
				chanOS <- ToPeerUpdate(peerList)
			}
		}
	}
}
