package localsingle

const (
	N_FLOORS  = 4
	N_BUTTONS = 3
)

type elevatorBehaviour int

const (
	BehaviourIdle elevatorBehaviour = iota
	BehaviourDoorOpen
	BehaviourMoving
)

type direction int

const (
	DirDown direction = -1
	DirStop direction = 0
	DirUp   direction = 1
)

type buttonType int

const (
	BtnHallUp buttonType = iota
	BtnHallDown
	BtnCab
)

type elevatorState struct {
	floor     int
	direction direction
	behaviour elevatorBehaviour
}

type LocalSingleElevator struct {
	state      elevatorState
	requests   [N_FLOORS][N_BUTTONS]bool
	obstructed bool
}

type directionBehaviourPair struct {
	direction direction
	behaviour elevatorBehaviour
}

func elevatorBehaviourToString(eb elevatorBehaviour) string {
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

func makeUninitializedElevator() LocalSingleElevator {
	elevator := LocalSingleElevator{
		state: elevatorState{floor: -1,
			direction: DirStop,
			behaviour: BehaviourIdle,
		},
	}
	return elevator
}

type commandType int

const (
	setMotorDirection commandType = iota
	setDoorOpenLamp
	setFloorIndicator
	setButtonLamp
	resetDoorTimer
	sendClearedOrders
)

type command struct {
	_type commandType
	value any
}

type buttonLampArgs struct {
	Floor int
	Btn   buttonType
	Value bool
}

type Order struct {
	Floor  int
	Button buttonType
}

// func PrintElevator(elevator LocalSingleElevator) {
// 	fmt.Printf("  +--------------------+\n")
// 	fmt.Printf(
// 		"  |floor = %-2d          |\n"+
// 			"  |direction  = %-12.12s|\n"+
// 			"  |behav = %-12.12s|\n",
// 		elevator.state.floor,
// 		DirectionToString(elevator.state.direction),
// 		ElevatorBehaviourToString(elevator.state.behaviour),
// 	)
// 	fmt.Printf("  +--------------------+\n")
// 	fmt.Printf("  |  | up  | dn  | cab |\n")
// 	for f := N_FLOORS - 1; f >= 0; f-- {
// 		fmt.Printf("  | %d", f)
// 		for btn := 0; btn < N_BUTTONS; btn++ {
// 			if (f == N_FLOORS-1 && btn == int(elevio.BT_HallUp)) ||
// 				(f == 0 && btn == int(elevio.BT_HallDown)) {
// 				fmt.Printf("|     ")
// 			} else {
// 				if elevator.requests[f][btn] != false {
// 					fmt.Print("|  #  ")
// 				} else {
// 					fmt.Print("|  -  ")
// 				}
// 			}
// 		}
// 		fmt.Printf("|\n")
// 	}
// 	fmt.Printf("  +--------------------+\n")
// }
