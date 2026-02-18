package peermonitor

import (
	ordersync "Project/OrderSync"
	"time"
)

// core PeerMonitor

func HandleHeartbeats(peerList []Peer, hb ordersync.NetMsg, now time.Time) ([]Peer, bool) {
	// Update or create peer, set Alive, update cab calls, set LastSeen
}

func CheckTimeouts(peerList []Peer, now time.Time, timeout time.Duration) ([]Peer, bool) {
	// Mark peers as Dead if LastSeen + timeout < now
}

func ToPeerUpdate(peerList []Peer) PeerUpdate {
	//Convert internal state to output format
}

func SendhbTx(hbTx chan<- RecoveryMsg, message RecoveryMsg) {

}
