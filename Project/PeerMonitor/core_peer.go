package peermonitor

import (
	"time"
)

// core PeerMonitor

func HandleHeartbeats(peerList []Peer, msg NetMsg, now time.Time) ([]Peer, bool) {
	// Update or create peer, set Alive , set LastSeen
	changed := false
	PeerID := msg.ElevID

	index := findPeerIndex(peerList, PeerID)
	if index == -1 { //-1 if peer not in Peerlist
		peerList = append(peerList, Peer{
			ID:         PeerID,
			PeerStatus: StatusAlive,
			LastSeen:   now,
		})
		return peerList, true
	}

	//Peer exists is Dead -> Alive,
	if peerList[index].PeerStatus != StatusAlive {
		peerList[index].PeerStatus = StatusAlive
		changed = true
	}

	peerList[index].LastSeen = now

	return peerList, changed

}

func CheckTimeouts(peerList []Peer, now time.Time, timeout time.Duration) ([]Peer, bool) {
	// Mark peers as Dead if LastSeen + timeout < now
	changed := false

	for index := range peerList {
		peer := peerList[index]
		timeSinceLastSeen := now.Sub(peer.LastSeen)

		if peer.PeerStatus == StatusAlive && timeSinceLastSeen > timeout { //check for timeout
			peerList[index].PeerStatus = StatusDead //declared dead
			changed = true
		}
	}

	return peerList, changed
}

func ToPeerUpdate(peerList []Peer) PeerUpdate { // makes a copy of peerList before sending to avoid sharing mutable state between goroutines
	out := make([]Peer, len(peerList))
	copy(out, peerList)
	return PeerUpdate{Peers: out}
}

func findPeerIndex(peers []Peer, id ElevID) int { //finds ID of peer
	for index := range peers {
		if peers[index].ID == id {
			return index
		}
	}
	return -1
}
