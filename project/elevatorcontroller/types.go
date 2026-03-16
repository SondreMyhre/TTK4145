package elevatorcontroller

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
	Floor      int
	Direction  Direction
	Behaviour  ElevatorBehaviour
	Obstructed bool
	MotorStuck bool
}

type RequestMatrix = [N_FLOORS][N_BUTTONS]bool

type elevator struct {
	state    ElevatorState
	requests [N_FLOORS][N_BUTTONS]bool
}

type directionBehaviourPair struct {
	direction Direction
	behaviour ElevatorBehaviour
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

type commandType int

const (
	setMotorDirection commandType = iota
	setDoorOpenLamp
	setFloorIndicator
	setButtonLamp
	resetDoorTimer
	sendClearedOrders
	sendLocalState
)

type command struct {
	cmdType commandType
	value   any
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
