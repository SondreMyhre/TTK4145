package ordersync

import (
	elevio "project/elevio"
	localsingle "project/localsingleelevator"
)

const (
	N_FLOORS = 4
	N_HALL   = 2
)

type OrderStatus int

const (
	Inactive OrderStatus = iota
	Pending
	Confirmed
	Assigned
)

type ElevID string

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

type CabCallsMap map[ElevID][N_FLOORS]bool

type NetMsg struct {
	SenderID        ElevID
	HallOrderMatrix HallOrderMatrix
	CabCalls        CabCallsMap
	SenderState     localsingle.ElevatorState
}

type HallOrderMatrix [N_FLOORS][N_HALL]OrderMatrixEntry

type OrderMatrixEntry struct {
	Status           OrderStatus
	AssignedElevator ElevID
	Version          int
}

type buttonLampArgs struct {
	Floor  int
	Button elevio.ButtonType
	Value  bool
}

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

type OrderLocation struct {
	Floor  int
	Button elevio.ButtonType
	Entry  OrderMatrixEntry
}
