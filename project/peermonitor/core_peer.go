package peermonitor

import (
	"time"
)

// core PeerMonitor

func sendHeartbeats(peerID string, heartBeatTx chan<- HeartBeat, hBticker *time.Ticker) {

	for range hBticker.C {
		heartBeat := HeartBeat{SenderID: ElevID(peerID)}
		heartBeatTx <- heartBeat
	}
}

func HandleHeartbeats(peerList []Peer, heartBeat HeartBeat, now time.Time) ([]Peer, bool) {
	// Update or create peer, set Alive , set LastSeen
	changed := false
	peerID := heartBeat.SenderID

	index := findPeerIndex(peerList, peerID)
	if index == -1 { //-1 if peer not in Peerlist
		peerList = append(peerList, Peer{
			ID:         peerID,
			PeerStatus: StatusAlive,
			lastSeen:   now,
		})
		return peerList, true
	}

	//Peer exists is Dead -> Alive,
	if peerList[index].PeerStatus != StatusAlive {
		peerList[index].PeerStatus = StatusAlive
		changed = true
	}

	peerList[index].lastSeen = now

	return peerList, changed

}

func CheckTimeouts(peerList []Peer, now time.Time, timeout time.Duration) ([]Peer, bool) {
	// Mark peers as Dead if LastSeen + timeout < now
	changed := false

	for index := range peerList {
		peer := peerList[index]
		timeSinceLastSeen := now.Sub(peer.lastSeen)

		if peer.PeerStatus == StatusAlive && timeSinceLastSeen > timeout { //check for timeout
			peerList[index].PeerStatus = StatusDead //declared dead
			changed = true
		}
	}

	return peerList, changed
}

// makes a copy of peerList before sending to avoid sharing mutable state between goroutines

func ToPeerUpdate(peerList []Peer) PeerMsg {
	out := make([]PeerUpdate, len(peerList))
	for index, peer := range peerList {
		out[index] = PeerUpdate{
			ID:         peer.ID,
			PeerStatus: peer.PeerStatus,
		}
	}

	return PeerMsg{Peers: out}
}

func findPeerIndex(peers []Peer, id ElevID) int { //finds ID of peer
	for index := range peers {
		if peers[index].ID == id {
			return index
		}
	}
	return -1
}
