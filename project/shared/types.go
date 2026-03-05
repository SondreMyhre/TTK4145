package shared

const (
	N_FLOORS  = 4
	N_BUTTONS = 3
	N_HALL	  = 2
)

type HallRequests [N_FLOORS][N_HALL]bool
type CabRequests  [N_FLOORS]bool
type Floor int
type ElevID int


type Direction int

const (
	DirDown Direction = -1
	DirStop Direction = 0
	DirUp   Direction = 1
)


type ElevatorBehaviour int

const (
	BehaviourIdle ElevatorBehaviour = iota
	BehaviourDoorOpen
	BehaviourMoving
)

type ElevatorState struct {
	Floor      int
	Direction  Direction
	Behaviour  ElevatorBehaviour
	Obstructed bool
}

type ButtonType int

const (
	BtnHallUp ButtonType = iota
	BtnHallDown
	BtnCab
)