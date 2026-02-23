package peermonitor

import (
	shared "Project/sharedtypes"
	"time"
)


// types

type ElevID = shared.ElevID
type NetMsg = shared.NetMsg

// type MsgType int

// const(
// 	NetMsg MsgType = iota //t
// 	PeerStatusMsg 
// )

type PeerStatus int

const ( // Status is Dead/Alive
	StatusDead PeerStatus = iota
	StatusAlive
)



type Peer struct { //Peer representsa network peer
	ID             ElevID      // elevator ID is based on the port i comes from since each IP is the same
	PeerStatus         PeerStatus             // Dead or alive
	LastSeen       time.Time          // When we last received a heartbeat from this peer
}

type PeerUpdate struct {
	Peers []Peer // Current list of all peers with their status
}

type PeerConfig struct {
	Timeout time.Duration // How long before a peer is declared dead
	TickInterval   time.Duration // internall ticker
}


