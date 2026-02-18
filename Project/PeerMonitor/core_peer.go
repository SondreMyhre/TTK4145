package peermonitor

import "time"

// core PeerMonitor



func HandleHeartbeats(state PeerState, hb HeartbeatMsg, now time.Time) PeerState{
	// Update or create peer, set Alive, update cab calls, set LastSeen
}

func CheckTimeouts(state PeerState, now time.Time, timeout time.Duration) PeerState {
	// Mark peers as Dead if LastSeen + timeout < now
}

func ToPeerUpdate(state PeerState) PeerUpdate{
	//Convert internal state to output format
}





