package ordersync

import (
	localsingle "project/localsingleelevator"
)

const (
	N_FLOORS  = localsingle.N_FLOORS
	N_HALL    = 2
	N_BUTTONS = localsingle.N_BUTTONS
	BT_CAB    = 2
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

type NetMsg struct {
	SenderID        ElevID
	HallOrderMatrix HallOrderMatrix
	CabCalls        CabCallsMap
	SenderState     localsingle.ElevatorState
}

type WorldviewMsg struct {
	HallRequests HallRequests
	CabRequests  CabCallsMap
	PeerStates   map[ElevID]localsingle.ElevatorState
	Peers        []Peer
}

type worldviewState struct {
	hallOrderMatrix HallOrderMatrix
	cabRequests     CabCallsMap
	pendingCabCalls [N_FLOORS]bool
	peerList        []Peer
	localState      localsingle.ElevatorState
}

// -------------------------------------------------------------

type commandType int

const (
	broadcastNetMessage commandType = iota
	setButtonLamp
)

type command struct {
	_type commandType
	value any
}

type buttonLampArgs struct {
	Floor  int
	Button int
	Value  bool
}
