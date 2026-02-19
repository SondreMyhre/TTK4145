package sharedtypes

const (
	N_FLOORS  = 4
	N_BUTTONS = 3
	N_HALL    = 2
	N_CAB     = 1
)

type ElevID int
type CabCallsMap [][]bool;

type HallOrderMatrix [N_FLOORS][N_HALL]orderMatrixEntry
type orderMatrixEntry struct {
	orderStatus      orderStatus
	assignedElevator ElevID
	version          int
}

type orderStatus int



type LocalCabCalls [N_FLOORS]bool

type NetMsg struct {
	ElevID          ElevID
	HallOrderMatrix HallOrderMatrix
	CabCalls        CabCallsMap
}
