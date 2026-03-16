package localsingle

func (elevator *elevator) requestsAbove() bool {
	for floor := elevator.state.Floor + 1; floor < N_FLOORS; floor++ {
		for btn := range N_BUTTONS {
			if elevator.requests[floor][btn] {
				return true
			}
		}
	}
	return false
}

func (elevator *elevator) requestsBelow() bool {
	for floor := range elevator.state.Floor {
		for btn := range N_BUTTONS {
			if elevator.requests[floor][btn] {
				return true
			}
		}
	}
	return false
}

func (elevator *elevator) requestsHere() bool {
	for btn := range N_BUTTONS {
		if elevator.requests[elevator.state.Floor][btn] {
			return true
		}
	}
	return false
}

func (elevator *elevator) chooseDirection() directionBehaviourPair {
	switch elevator.state.Direction {
	case DirUp:
		if elevator.requestsAbove() {
			return directionBehaviourPair{DirUp, BehaviourMoving}
		} else if elevator.requestsHere() {
			return directionBehaviourPair{DirStop, BehaviourDoorOpen}
		} else if elevator.requestsBelow() {
			return directionBehaviourPair{DirDown, BehaviourMoving}
		} else {
			return directionBehaviourPair{DirStop, BehaviourIdle}
		}
	case DirDown:
		if elevator.requestsBelow() {
			return directionBehaviourPair{DirDown, BehaviourMoving}
		} else if elevator.requestsHere() {
			return directionBehaviourPair{DirStop, BehaviourDoorOpen}
		} else if elevator.requestsAbove() {
			return directionBehaviourPair{DirUp, BehaviourMoving}
		} else {
			return directionBehaviourPair{DirStop, BehaviourIdle}
		}
	case DirStop:
		if elevator.requestsHere() {
			return directionBehaviourPair{DirStop, BehaviourDoorOpen}
		} else if elevator.requestsAbove() {
			return directionBehaviourPair{DirUp, BehaviourMoving}
		} else if elevator.requestsBelow() {
			return directionBehaviourPair{DirDown, BehaviourMoving}
		} else {
			return directionBehaviourPair{DirStop, BehaviourIdle}
		}
	default:
		return directionBehaviourPair{DirStop, BehaviourIdle}
	}
}

func (elevator *elevator) shouldStop() bool {
	switch elevator.state.Direction {
	case DirDown:
		return elevator.requests[elevator.state.Floor][BtnHallDown] ||
			elevator.requests[elevator.state.Floor][BtnCab] ||
			!elevator.requestsBelow()
	case DirUp:
		return elevator.requests[elevator.state.Floor][BtnHallUp] ||
			elevator.requests[elevator.state.Floor][BtnCab] ||
			!elevator.requestsAbove()
	default:
		return true
	}
}

func (elevator *elevator) clearAtCurrentFloor() []Order {
	var clearedOrders []Order

	if elevator.requests[elevator.state.Floor][BtnCab] {
		elevator.requests[elevator.state.Floor][BtnCab] = false
		clearedOrders = append(clearedOrders, Order{elevator.state.Floor, BtnCab})
	}

	switch elevator.state.Direction {
	case DirUp:
		if elevator.requests[elevator.state.Floor][BtnHallUp] {
			elevator.requests[elevator.state.Floor][BtnHallUp] = false
			clearedOrders = append(clearedOrders, Order{elevator.state.Floor, BtnHallUp})
		}

		if !elevator.requestsAbove() && !elevator.requests[elevator.state.Floor][BtnHallUp] && elevator.requests[elevator.state.Floor][BtnHallDown] {
			elevator.requests[elevator.state.Floor][BtnHallDown] = false
			clearedOrders = append(clearedOrders, Order{elevator.state.Floor, BtnHallDown})
		}

	case DirDown:
		if elevator.requests[elevator.state.Floor][BtnHallDown] {
			elevator.requests[elevator.state.Floor][BtnHallDown] = false
			clearedOrders = append(clearedOrders, Order{elevator.state.Floor, BtnHallDown})
		}

		if !elevator.requestsBelow() && !elevator.requests[elevator.state.Floor][BtnHallDown] && elevator.requests[elevator.state.Floor][BtnHallUp] {
			elevator.requests[elevator.state.Floor][BtnHallUp] = false
			clearedOrders = append(clearedOrders, Order{elevator.state.Floor, BtnHallUp})
		}

	default:
		if elevator.requests[elevator.state.Floor][BtnHallUp] {
			elevator.requests[elevator.state.Floor][BtnHallUp] = false
			clearedOrders = append(clearedOrders, Order{elevator.state.Floor, BtnHallUp})
		}

		if elevator.requests[elevator.state.Floor][BtnHallDown] {
			elevator.requests[elevator.state.Floor][BtnHallDown] = false
			clearedOrders = append(clearedOrders, Order{elevator.state.Floor, BtnHallDown})
		}
	}

	return clearedOrders
}
