package localsingle

import (
	elevio "TTK4145/ElevIO"
	"fmt"
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
	floor              int
	direction          Direction
	behaviour          ElevatorBehaviour
}

type LocalSingleElevator struct {
	state 			   ElevatorState
	requests           [N_FLOORS][N_BUTTONS]bool
	doorOpenDuration_s float64

	// dropper config

}

func eb_toString(eb ElevatorBehaviour) string {
	switch eb {
	case BehaviourIdle:
		return "EB_Idle"
	case BehaviourDoorOpen:
		return "EB_DoorOpen"
	case BehaviourMoving:
		return "EB_Moving"
	default:
		return "EB_UNDEFINED"

	}
}

func direction_toString(direction Direction) string {
	switch direction {
	case DirUp:
		return "D_Up"
	case DirDown:
		return "D_Down"
	case DirStop:
		return "D_Stop"
	default:
		return "D_UNDEFINED"

	}
}

func elevator_print(elevator LocalSingleElevator) {
	fmt.Printf("  +--------------------+\n")
	fmt.Printf(
		"  |floor = %-2d          |\n"+
			"  |direction  = %-12.12s|\n"+
			"  |behav = %-12.12s|\n",
		elevator.state.floor,
		direction_toString(elevator.state.direction),
		eb_toString(elevator.state.behaviour),
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

func elevator_uninitialized() LocalSingleElevator {
	return LocalSingleElevator{
		state: ElevatorState{floor:              -1,
							 direction:          DirStop,
							 behaviour:          BehaviourIdle,
							},
		doorOpenDuration_s: 3.0,
	}
}
