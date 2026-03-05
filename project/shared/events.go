package shared

type PeerHealth struct {
	ID ElevID
	Alive bool
}

type PeerUpdate struct {
	Peers []PeerHealth
}

type ElevatorState struct {
	Floor      int
	Direction  Direction
	Behaviour  ElevatorBehaviour
	Faulted bool
}