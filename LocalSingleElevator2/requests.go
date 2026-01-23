package localsingle2

// DirnBehaviourPair holds a direction and behaviour combination.
type DirnBehaviourPair struct {
	Direction Direction
	Behaviour Behaviour
}

// requestsAbove checks if there are any requests above the current floor.
func requestsAbove(e Elevator) bool {
	for f := e.Floor + 1; f < NumFloors; f++ {
		for btn := 0; btn < NumButtons; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

// requestsBelow checks if there are any requests below the current floor.
func requestsBelow(e Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < NumButtons; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

// requestsHere checks if there are any requests at the current floor.
func requestsHere(e Elevator) bool {
	if e.Floor < 0 || e.Floor >= NumFloors {
		return false
	}
	for btn := 0; btn < NumButtons; btn++ {
		if e.Requests[e.Floor][btn] {
			return true
		}
	}
	return false
}

// ChooseDirection determines the next direction and behaviour based on current requests.
func ChooseDirection(e Elevator) DirnBehaviourPair {
	switch e.Direction {
	case DirUp:
		if requestsAbove(e) {
			return DirnBehaviourPair{DirUp, BehaviourMoving}
		}
		if requestsHere(e) {
			return DirnBehaviourPair{DirDown, BehaviourDoorOpen}
		}
		if requestsBelow(e) {
			return DirnBehaviourPair{DirDown, BehaviourMoving}
		}
		return DirnBehaviourPair{DirStop, BehaviourIdle}

	case DirDown:
		if requestsBelow(e) {
			return DirnBehaviourPair{DirDown, BehaviourMoving}
		}
		if requestsHere(e) {
			return DirnBehaviourPair{DirUp, BehaviourDoorOpen}
		}
		if requestsAbove(e) {
			return DirnBehaviourPair{DirUp, BehaviourMoving}
		}
		return DirnBehaviourPair{DirStop, BehaviourIdle}

	case DirStop:
		// Only one request expected in stop case; checking up or down first is arbitrary.
		if requestsHere(e) {
			return DirnBehaviourPair{DirStop, BehaviourDoorOpen}
		}
		if requestsAbove(e) {
			return DirnBehaviourPair{DirUp, BehaviourMoving}
		}
		if requestsBelow(e) {
			return DirnBehaviourPair{DirDown, BehaviourMoving}
		}
		return DirnBehaviourPair{DirStop, BehaviourIdle}

	default:
		return DirnBehaviourPair{DirStop, BehaviourIdle}
	}
}

// ShouldStop determines if the elevator should stop at the current floor.
func ShouldStop(e Elevator) bool {
	switch e.Direction {
	case DirDown:
		return e.Requests[e.Floor][ButtonHallDown] ||
			e.Requests[e.Floor][ButtonCab] ||
			!requestsBelow(e)

	case DirUp:
		return e.Requests[e.Floor][ButtonHallUp] ||
			e.Requests[e.Floor][ButtonCab] ||
			!requestsAbove(e)

	case DirStop:
		return true

	default:
		return true
	}
}

// ShouldClearImmediately determines if a request should be cleared immediately
// without adding it to the request queue.
func ShouldClearImmediately(e Elevator, btnFloor int, btnType ButtonType) bool {
	return e.Floor == btnFloor &&
		((e.Direction == DirUp && btnType == ButtonHallUp) ||
			(e.Direction == DirDown && btnType == ButtonHallDown) ||
			e.Direction == DirStop ||
			btnType == ButtonCab)
}

// ClearAtCurrentFloor clears appropriate requests at the current floor
// based on the elevator's direction.
func ClearAtCurrentFloor(e Elevator) Elevator {
	e.Requests[e.Floor][ButtonCab] = false

	switch e.Direction {
	case DirUp:
		if !requestsAbove(e) && !e.Requests[e.Floor][ButtonHallUp] {
			e.Requests[e.Floor][ButtonHallDown] = false
		}
		e.Requests[e.Floor][ButtonHallUp] = false

	case DirDown:
		if !requestsBelow(e) && !e.Requests[e.Floor][ButtonHallDown] {
			e.Requests[e.Floor][ButtonHallUp] = false
		}
		e.Requests[e.Floor][ButtonHallDown] = false

	case DirStop:
		e.Requests[e.Floor][ButtonHallUp] = false
		e.Requests[e.Floor][ButtonHallDown] = false
	}

	return e
}
