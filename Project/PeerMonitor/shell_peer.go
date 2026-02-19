package peermonitor

import (
	ordersync "Project/OrderSync"
	"time"
)

// Shell PeerMonitor

func Run(cfg PeerConfig, hbRx <-chan ordersync.NetMsg, chanOS chan<- PeerUpdate) {
	var peerList []Peer

	ticker := time.NewTicker(50 * time.Millisecond) //creates ticker* struct, ticker.C is channel
	defer ticker.Stop() //runs ticker while function is running

	for {
		select {
		case hb := <-hbRx:
			// if ok != nil{ // se om kanal er åpen??
			// 	return 
			// }
			var changed bool

			peerList, changed = HandleHeartbeats(peerList, hb, time.Now()) //looks for updates and sets changed to true/false
			if changed {
				chanOS <- ToPeerUpdate(peerList)
			}
		case <-ticker.C: //C is channel if ticker
			var timeoutChanged bool
			peerList, timeoutChanged = CheckTimeouts(peerList,time.Now() , cfg.Timeout) //looks for updates and sets changed to true/false
			if timeoutChanged {
				chanOS <- ToPeerUpdate(peerList)
			}
		}
	}
}


