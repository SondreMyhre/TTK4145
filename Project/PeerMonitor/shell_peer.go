package peermonitor

import(
	"time"
	ordersync "Project/OrderSync"

) 

// Shell PeerMonitor





func Run(cfg PeerConfig, hbRx <-chan ordersync.NetMsg, chanOS chan<- PeerUpdate, RecoveryTx chan<- RecoveryMsg) {
	var peerList []Peer  

	
    for {
        select {
        case hb := <-hbRx:
			var changed bool
			now := time.Now()


            peerList, changed = HandleHeartbeats(peerList, hb, now)
			if changed{
				chanOS <- ToPeerUpdate(peerList)
			}
			var timeoutChanged bool
			peerList, timeoutChanged = CheckTimeouts(peerList, now, cfg.Timeout)
			if timeoutChanged{
				chanOS <- ToPeerUpdate(peerList)
        	}
		}
    }
}

