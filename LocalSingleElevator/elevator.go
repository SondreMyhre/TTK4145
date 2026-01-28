package localsingle

import (
	elevio "TTK4145/ElevIO"
	"fmt"
	"time"
)

const (
	N_FLOORS  = 4
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

func PrintElevator(elevator LocalSingleElevator) {
	fmt.Printf("  +--------------------+\n")
	fmt.Printf(
		"  |floor = %-2d          |\n"+
			"  |direction  = %-12.12s|\n"+
			"  |behav = %-12.12s|\n",
		elevator.state.floor,
		DirectionToString(elevator.state.direction),
		ElevatorBehaviourToString(elevator.state.behaviour),
	)
	fmt.Printf("  +--------------------+\n")
	fmt.Printf("  |  | up  | dn  | cab |\n")
	for f := N_FLOORS - 1; f >= 0; f-- {
		fmt.Printf("  | %d", f)
		for btn := 0; btn < N_BUTTONS; btn++ {
			if (f == N_FLOORS-1 && btn == int(elevio.BT_HallUp)) ||
				(f == 0 && btn == int(elevio.BT_HallDown)) {
				fmt.Printf("|     ")
			} else {
				if elevator.requests[f][btn] != false {
					fmt.Print("|  #  ")
				} else {
					fmt.Print("|  -  ")
				}
			}
		}
		fmt.Printf("|\n")
	}
	fmt.Printf("  +--------------------+\n")
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

func (elevator *LocalSingleElevator) InitDoorTimer() {
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
