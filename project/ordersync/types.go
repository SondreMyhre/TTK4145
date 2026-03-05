package ordersync

import (
	elevio "project/elevio"
	localsingle "project/localsingleelevator"
)

const (
	N_FLOORS  = 4
	N_HALL    = 2
	N_BUTTONS = 3
)

type ElevID string

type OrderStatus int

const (
	Inactive OrderStatus = iota
	Pending
	Confirmed
)

type OrderMatrixEntry struct {
	Status  OrderStatus
	Version int
}

type HallOrderMatrix [N_FLOORS][N_HALL]OrderMatrixEntry

type CabCallsMap map[ElevID][N_FLOORS]bool

type HallRequests [N_FLOORS][N_HALL]bool
type CabRequests [N_FLOORS]bool

type PeerStatus int

const (
	StatusDead PeerStatus = iota
	StatusAlive
)

type Peer struct {
	ID         ElevID
	PeerStatus PeerStatus
	state      localsingle.ElevatorState
}

type PeerUpdate struct {
	ID         ElevID
	PeerStatus PeerStatus
}

type PeerMsg struct {
	Peers []PeerUpdate
}

type NetMsg struct {
	SenderID        ElevID
	HallOrderMatrix HallOrderMatrix
	CabCalls        CabCallsMap
	SenderState     localsingle.ElevatorState
}

type Worldview struct {
	HallRequests HallRequests
	CabRequests  map[ElevID]CabRequests
	PeerStates   map[ElevID]localsingle.ElevatorState
	Peer         []Peer
}

// -------------------------------------------------------------

type commandType int

const (
	sendOrderToLocal commandType = iota
	broadcastNetMessage
	setButtonLamp
)

type command struct {
	_type commandType
	value any
}

type OrderLocation struct {
	Floor  int
	Button elevio.ButtonType
	Entry  OrderMatrixEntry
}
