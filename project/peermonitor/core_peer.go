package peermonitor

import (
	ordersync "project/ordersync"
	"time"
)

// core PeerMonitor

func HandleHeartbeats(peerList []Peer, msg ordersync.NetMsg, now time.Time) ([]Peer, bool) {
	// Update or create peer, set Alive , set LastSeen
	changed := false
	PeerID := msg.SenderID

	i := findPeerIndex(peerList, PeerID)
	if i == -1 { //-1 if peer not in Peerlist
		peerList = append(peerList, Peer{
			ID:             PeerID,
			Status:         Alive,
			LastSeen:       now,
		})
		return peerList, true
	}

	//Peer exists is Dead -> Alive,
	if peerList[i].Status != Alive {
		peerList[i].Status = Alive
		changed = true
	}

	peerList[i].LastSeen = now 

	return peerList, changed

}

func CheckTimeouts(peerList []Peer, now time.Time, timeout time.Duration) ([]Peer, bool) {
	// Mark peers as Dead if LastSeen + timeout < now
	changed := false

	for i := range peerList {
		if peerList[i].Status == Alive && now.Sub(peerList[i].LastSeen) > timeout { //check for timeout
			peerList[i].Status = Dead
			changed = true
		}
	}
	return peerList, changed
}

func ToPeerUpdate(peerList []Peer) []ordersync.Peer { // makes a copy of peerList before sending to avoid sharing mutable state between goroutines
	out := make([]ordersync.Peer, len(peerList))
	for i, peer := range peerList {
		out[i] = ordersync.Peer{
			ID: peer.ID,
			Status: ordersync.PeerStatus(peer.Status),
		}
	}
	return out
}

func findPeerIndex(peers []Peer, id ordersync.ElevID) int { //finds ID of peer
	for i := range peers {
		if peers[i].ID == id {
			return i
		}
	}
	return -1
}

