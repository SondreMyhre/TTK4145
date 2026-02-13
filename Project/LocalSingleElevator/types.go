package localsingle

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

type ButtonType int

const (
	BtnHallUp ButtonType = iota
	BtnHallDown
	BtnCab
)

type ElevatorState struct {
	floor     int
	direction Direction
	behaviour ElevatorBehaviour
}

type LocalSingleElevator struct {
	state      ElevatorState
	requests   [N_FLOORS][N_BUTTONS]bool
	obstructed bool
}

type DirectionBehaviourPair struct {
	direction Direction
	behaviour ElevatorBehaviour
}

func elevatorBehaviourToString(eb ElevatorBehaviour) string {
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
		state: ElevatorState{floor: -1,
			direction: DirStop,
			behaviour: BehaviourIdle,
		},
	}
	return elevator
}

type CommandType int

const (
	setMotorDirection CommandType = iota
	setDoorOpenLamp
	setFloorIndicator
	setButtonLamp
	resetDoorTimer
	sendClearedOrders
)

type Command struct {
	_type CommandType
	value any
}

type ButtonLampArgs struct {
	Floor int
	Btn   ButtonType
	Value bool
}

type Order struct {
	Floor  int
	Button ButtonType
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