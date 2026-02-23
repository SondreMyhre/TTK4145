package peermonitor

import (
	ordersync "project/ordersync"
	"time"
)


// types


type Status int

const ( // Status is Dead/Alive
	Dead Status = iota
	Alive
)

type Peer struct { //Peer representsa network peer
	ID             ordersync.ElevID      // elevator ID is based on the port i comes from since each IP is the same
	Status         Status             //Dead or alive
	LastSeen       time.Time          // When we last received a heartbeat from this peer
}

type PeerUpdate struct {
	Peers []ordersync.Peer // Current list of all peers with their status
}

type PeerConfig struct {
	Timeout time.Duration // How long before a peer is declared dead
}
