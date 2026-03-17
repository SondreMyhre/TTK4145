package peermonitor

import (
	"time"
)

func HandleHeartbeats(peerList []Peer, heartBeat HeartBeat, now time.Time) ([]Peer, bool) {
	changed := false
	peerID := heartBeat.SenderID

	index := findPeerIndex(peerList, peerID)
	if index == -1 { 
		peerList = append(peerList, Peer{
			ID:         peerID,
			PeerStatus: StatusAlive,
			lastSeen:   now,
		})
		return peerList, true
	}

	if peerList[index].PeerStatus != StatusAlive {
		peerList[index].PeerStatus = StatusAlive
		changed = true
	}

	peerList[index].lastSeen = now

	return peerList, changed

}

func CheckTimeouts(peerList []Peer, now time.Time, timeout time.Duration) ([]Peer, bool) {
	changed := false

	for index := range peerList {
		peer := peerList[index]
		timeSinceLastSeen := now.Sub(peer.lastSeen)

		if peer.PeerStatus == StatusAlive && timeSinceLastSeen > timeout { 
			peerList[index].PeerStatus = StatusDead 
			changed = true
		}
	}

	return peerList, changed
}

func ToPeerUpdate(peerList []Peer) PeerMsg {
	out := make([]PeerUpdate, len(peerList))
	for index, peer := range peerList {
		out[index] = PeerUpdate{
			ID:         peer.ID,
			PeerStatus: peer.PeerStatus,
		}
	}

	return out
}

func findPeerIndex(peers []Peer, id ElevID) int { 
	for index := range peers {
		if peers[index].ID == id {
			return index
		}
	}
	return -1
}
