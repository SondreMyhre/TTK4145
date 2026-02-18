package peermonitor

import "time"

const N_Floors = 4

// types

type ElevID string //Elev ID uniquely identifies a Peer

type Status int

const ( // Status is Dead/Alive
	Dead Status = iota
	Alive
)

type Peer struct { //Peer representsa network peer
	ID             ElevID         // elevator ID is based on the port i comes from since each IP is the same
	Status         Status         //Dead or alive
	BackupCabCalls [N_Floors]bool //each elevator keeps backup of others' cab calls
	LastSeen       time.Time      // When we last received a heartbeat from this peer
}

type RecoveryMsg struct { //HeartbeatMsg is recieved from other alive Peers
	SenderID ElevID
	CabCalls [N_Floors]bool // Current cab calls to be backed up by others
}

type PeerUpdate struct {
	Peers []Peer // Current list of all peers with their status
}

type PeerConfig struct {
	Timeout    time.Duration // How long before a peer is declared dead
	TickPeriod time.Duration // 0 => expects external tick when testing, >0 => creates internal ticker in production
}

// type PeerInputs struct {
// 	Heartbeat <-chan HeartbeatMsg // Received from TransportUDP Rx
// 	Tick      <-chan time.Time    // External tick, checks for dead peers //For Testing
// }

// type PeerOutputs struct {
// 	Update              chan<- PeerUpdate // Sends updated list to OrderSync
// 	TransmitBackupCalls chan<- HeartbeatMsg     // Sends backup cab calls
// }
