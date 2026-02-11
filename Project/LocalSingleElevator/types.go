package localsingle

import (
	elevio "Project/ElevIO"
)

const (
	N_FLOORS  = 4	// TODO: gjøre dette mulig å sette med flags
	N_BUTTONS = 3
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
	floor     int
	direction Direction
	behaviour ElevatorBehaviour
}

type LocalSingleElevator struct {
	state     ElevatorState
	requests  [N_FLOORS][N_BUTTONS]bool
	doorTimer *time.Timer
	doorDur   time.Duration
}


type DirectionBehaviourPair struct {
	direction Direction
	behaviour ElevatorBehaviour
}

func ElevatorBehaviourToString(eb ElevatorBehaviour) string {
	switch eb {
	case BehaviourIdle:
		return "BehaviourIdle"
	case BehaviourDoorOpen:
		return "BehaviourDoorOpen"
	case BehaviourMoving:
		return "BehaviourMoving"
	default:
		return "BehaviourUndefined"

	}
}

func DirectionToString(direction Direction) string {
	switch direction {
	case DirUp:
		return "DirUp"
	case DirDown:
		return "DirDown"
	case DirStop:
		return "DirStop"
	default:
		return "DirUndefined"

	}
}

func DirectionToMotorDirection(direction Direction) elevio.MotorDirection {
	switch direction {
	case DirUp:
		return elevio.MD_Up
	case DirDown:
		return elevio.MD_Down
	case DirStop:
		return elevio.MD_Stop
	default:
		return elevio.MD_Stop
	}
}

func MakeUninitializedElevator() LocalSingleElevator {
	elevator := LocalSingleElevator{
		state: ElevatorState{floor: -1,
			direction: DirStop,
			behaviour: BehaviourIdle,
		},
	}
	elevator.InitDoorTimer()
	return elevator
}


type CommandType int

const (
	setMotor CommandType = iota
	setDoorLamp
	setFloorIndicator
	setButtonLamp
	startDoorTimer
	sendClearedOrders
)

type Command struct {
	_type CommandType
	value any
}

type ButtonLampArgs struct {
	Floor int
	Btn elevio.ButtonType
	Value bool
}

type Order struct {
	Floor int
	Button elevio.ButtonType
}


func (elevator *LocalSingleElevator) InitDoorTimer() { //Vurder å lage egen timer modul
	elevator.doorDur = 3 * time.Second

	// Lag timer, men start den "deaktivert"
	elevator.doorTimer = time.NewTimer(elevator.doorDur)
	if !elevator.doorTimer.Stop() {
		select {
		case <-elevator.doorTimer.C:
		default:
		}
	}
}

func (elevator *LocalSingleElevator) ResetDoorTimer() {
	if elevator.doorTimer == nil {
		elevator.InitDoorTimer()
	}
	if !elevator.doorTimer.Stop() {
		select {
		case <-elevator.doorTimer.C:
		default:
		}
	}
	elevator.doorTimer.Reset(elevator.doorDur)
}


