package elevatorcontroller

import(
	"time"
)

const (
	N_FLOORS  = 4
	N_BUTTONS = 3
)
const(
	doorOpenDuration     = 3 * time.Second
	motorWatchdogTimeout = 3500 * time.Millisecond
)
const (
	setMotorDirection commandType = iota
	setDoorOpenLamp
	setFloorIndicator
	setButtonLamp
	resetDoorTimer
	sendClearedOrders
	sendLocalState
)

type ElevatorBehaviour int
const (
	BehaviourIdle ElevatorBehaviour = iota
	BehaviourDoorOpen
	BehaviourMoving
)

type Direction int
const (
	DirDown Direction = -1
	DirStop Direction = 0
	DirUp   Direction = 1
)

type ElevatorState struct {
	Floor      int
	Direction  Direction
	Behaviour  ElevatorBehaviour
	Obstructed bool
	MotorStuck bool
}

type elevator struct {
	state    ElevatorState
	requests [N_FLOORS][N_BUTTONS]bool
}

type ButtonType int
const (
	BtnHallUp ButtonType = iota
	BtnHallDown
	BtnCab
)


type RequestMatrix = [N_FLOORS][N_BUTTONS]bool

type directionBehaviourPair struct {
	direction Direction
	behaviour ElevatorBehaviour
}

type commandType int
type command struct {
	kind  commandType
	value any
}

type buttonLampArgs struct {
	Floor int
	Btn   ButtonType
	Value bool
}

type Order struct {
	Floor  int
	Button ButtonType
}

func makeUninitializedElevator() elevator {
	elevator := elevator{
		state: ElevatorState{Floor: -1,
			Direction: DirStop,
			Behaviour: BehaviourIdle,
		},
	}
	return elevator
}

