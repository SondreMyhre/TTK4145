package ordersync

const (
	N_FLOORS  = 4
	N_BUTTONS = 3
	N_HALL    = 2
	N_CAB     = 1
)

type buttonType int

const (
	BtnHallUp buttonType = iota
	BtnHallDown
	BtnCab
)

type orderStatus int

const (
	inactive = iota
	local
	pending
	assigned
)

type ElevID int

type commandType int

const (
	sendOrderToLocal commandType = iota
	sendNetMsg
	setButtonLamp
)

type command struct {
	_type commandType
	value any
}

type localCabCalls [N_FLOORS]bool

type NetMsg struct {
	ElevID          ElevID
	HallOrderMatrix HallOrderMatrix
	CabCalls  map[ElevID]localCabCalls
}

type HallOrderMatrix [N_FLOORS][N_HALL]orderMatrixEntry

type orderMatrixEntry struct {
	orderStatus      orderStatus
	assignedElevator ElevID
	version          int
}

type localState int;

type buttonLampArgs int

const (
	Idle localState = iota
	Moving
	DoorOpen
)

type Status int

const (
	Dead Status = iota
	Alive
)


type peer struct {
	ID ElevID
	Status Status
}