package peermonitor

import (
	shared "Project/sharedtypes"
	"time"
)

// Shell PeerMonitor

func Run(cfg PeerConfig, hbRx <-chan shared.NetMsg, chanOS chan<- PeerUpdate) {
	var peerList []Peer

	ticker := time.NewTicker(50 * time.Millisecond) //creates ticker* struct, ticker.C is channel
	defer ticker.Stop()                             //runs ticker while function is running

	for {
		select {
		case msg,ok := <-hbRx:
			if !ok { // se om kanal er åpen, lukker dersom kanal er lukket??
				return
			}
			var changed bool

			peerList, changed = HandleHeartbeats(peerList, msg, time.Now()) //looks for updates and sets changed to true/false
			if changed {
				chanOS <- ToPeerUpdate(peerList)
			}
		case <-ticker.C: //C is channel if ticker
			var timeoutChanged bool
			peerList, timeoutChanged = CheckTimeouts(peerList, time.Now(), cfg.Timeout) //looks for updates and sets changed to true/false
			if timeoutChanged {
				chanOS <- ToPeerUpdate(peerList)
			}
		}
	}
}
