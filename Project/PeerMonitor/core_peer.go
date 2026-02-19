package peermonitor

import (
	ordersync "Project/OrderSync"
	"time"
)

// core PeerMonitor

func HandleHeartbeats(peerList []Peer, msg ordersync.NetMsg, now time.Time) ([]Peer, bool) {
	// Update or create peer, set Alive, update cab calls, set LastSeen
	changed := false

	//TODO: Tilpass disse to linjene til NetMsg
	PeerID := msg.ElevID
	NewCab := msg.BackupCabCalls

	i := findPeerIndex(peerList, PeerID)
	if i == -1{ //-1 if peer not in Peerlist
		peerList = append(peerList, Peer{
			ID: 	PeerID,
			Status: Alive,
			BackupCabCalls:	NewCab,
			LastSeen:	now,
		})
		return peerList, true
	}

	//Peer exists -> Alive, in case?
	if peerList[i].Status != Alive {
		peerList[i].Status = Alive
		changed = true
	}

	if !cabCallsEqual(peerList[i].BackupCabCalls, NewCab) { //checsks if cabcalls are equal
		peerList[i].BackupCabCalls = NewCab
		changed = true
	}
	peerList[i].LastSeen = now

	return peerList, changed

}

func CheckTimeouts(peerList []Peer, now time.Time, timeout time.Duration) ([]Peer, bool) {
	// Mark peers as Dead if LastSeen + timeout < now
	changed := false

	for i := range peerList{
		if peerList[i].Status == Alive && now.Sub(peerList[i].LastSeen) > timeout{
			peerList[i].Status = Dead
			changed = true
		}
	}
	return peerList, changed
}

func ToPeerUpdate(peerList []Peer) PeerUpdate {
	out := make([]Peer, len(peerList))
	copy(out, peerList)
	return PeerUpdate{Peers: out}
}


func findPeerIndex(peers []Peer, id ordersync.ElevID) int{ //finds ID
	for i := range peers{
		if peers[i].ID == id{
			return i
		}
	}
	return -1
}

func cabCallsEqual(a,b ordersync.CabCallArray) bool{ //checks for changes in backupCabcalls
	if len(a) != len(b){
		return false
	}

	for floor, aVal := range a{
		bVal, exists := b[floor]
		if !exists {
			return false
		}
		if bVal != aVal{
			return false
		}
	}
	return true
}



