package ordersync

import (
	"project/config"
	elevatorcontroller "project/elevatorcontroller"
)

const (
	N_FLOORS  = config.N_FLOORS
	N_HALL    = 2
	N_BUTTONS = elevatorcontroller.N_BUTTONS
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

type CabCalls struct {
	Map     map[ElevID][N_FLOORS]bool
	Version int
}

type HallRequests [N_FLOORS][N_HALL]bool

type PeerStatus int

const (
	StatusDead PeerStatus = iota
	StatusAlive
)

type Peer struct {
	ID         ElevID
	PeerStatus PeerStatus
	state      elevatorcontroller.ElevatorState
}

type PeerUpdate struct {
	ID         ElevID
	PeerStatus PeerStatus
}

type NetMsg struct {
	SenderID        ElevID
	HallOrderMatrix HallOrderMatrix
	CabCalls        CabCalls
	SenderState     elevatorcontroller.ElevatorState
}

type WorldviewMsg struct {
	HallRequests HallRequests
	CabRequests  map[ElevID][N_FLOORS]bool
	PeerStates   map[ElevID]elevatorcontroller.ElevatorState
	Peers        []Peer
}

type worldviewState struct {
	hallOrderMatrix HallOrderMatrix
	cabCalls        CabCalls
	peerList        []Peer
	localState      elevatorcontroller.ElevatorState
}

// -------------------------------------------------------------

type effectType int

const (
	broadcastNetMessage effectType = iota
	setButtonLamp
)

type effect struct {
	kind  effectType
	value any
}

type buttonLampArgs struct {
	Floor  int
	Button int
	Value  bool
}
