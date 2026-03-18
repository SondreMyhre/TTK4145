package elevatorcontroller

func requestsAbove(elevator elevator) bool {
	for floor := elevator.state.Floor + 1; floor < N_FLOORS; floor++ {
		for button := range N_BUTTONS {
			if elevator.requests[floor][button] {
				return true
			}
		}
	}
	return false
}

func requestsBelow(elevator elevator) bool {
	for floor := range elevator.state.Floor {
		for button := range N_BUTTONS {
			if elevator.requests[floor][button] {
				return true
			}
		}
	}
	return false
}

func requestsHere(elevator elevator) bool {
	for button := range N_BUTTONS {
		if elevator.requests[elevator.state.Floor][button] {
			return true
		}
	}
	return false
}

func chooseDirection(elevator elevator) directionBehaviourPair {
	switch elevator.state.Direction {
	case DirUp:
		if requestsAbove(elevator) {
			return directionBehaviourPair{DirUp, BehaviourMoving}
		} else if requestsHere(elevator) {
			return directionBehaviourPair{DirStop, BehaviourDoorOpen}
		} else if requestsBelow(elevator) {
			return directionBehaviourPair{DirDown, BehaviourMoving}
		} else {
			return directionBehaviourPair{DirStop, BehaviourIdle}
		}
	case DirDown:
		if requestsBelow(elevator) {
			return directionBehaviourPair{DirDown, BehaviourMoving}
		} else if requestsHere(elevator) {
			return directionBehaviourPair{DirStop, BehaviourDoorOpen}
		} else if requestsAbove(elevator) {
			return directionBehaviourPair{DirUp, BehaviourMoving}
		} else {
			return directionBehaviourPair{DirStop, BehaviourIdle}
		}
	case DirStop:
		if requestsHere(elevator) {
			return directionBehaviourPair{DirStop, BehaviourDoorOpen}
		} else if requestsAbove(elevator) {
			return directionBehaviourPair{DirUp, BehaviourMoving}
		} else if requestsBelow(elevator) {
			return directionBehaviourPair{DirDown, BehaviourMoving}
		} else {
			return directionBehaviourPair{DirStop, BehaviourIdle}
		}
	default:
		return directionBehaviourPair{DirStop, BehaviourIdle}
	}
}

func shouldStop(elevator elevator) bool {
	switch elevator.state.Direction {
	case DirDown:
		return elevator.requests[elevator.state.Floor][BtnHallDown] ||
			elevator.requests[elevator.state.Floor][BtnCab] ||
			!requestsBelow(elevator)
	case DirUp:
		return elevator.requests[elevator.state.Floor][BtnHallUp] ||
			elevator.requests[elevator.state.Floor][BtnCab] ||
			!requestsAbove(elevator)
	default:
		return true
	}
}

func clearAtCurrentFloor(elevator elevator) (elevator, []Order) {
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

		if !requestsAbove(elevator) && !elevator.requests[elevator.state.Floor][BtnHallUp] && elevator.requests[elevator.state.Floor][BtnHallDown] {
			elevator.requests[elevator.state.Floor][BtnHallDown] = false
			clearedOrders = append(clearedOrders, Order{elevator.state.Floor, BtnHallDown})
		}

	case DirDown:
		if elevator.requests[elevator.state.Floor][BtnHallDown] {
			elevator.requests[elevator.state.Floor][BtnHallDown] = false
			clearedOrders = append(clearedOrders, Order{elevator.state.Floor, BtnHallDown})
		}

		if !requestsBelow(elevator) && !elevator.requests[elevator.state.Floor][BtnHallDown] && elevator.requests[elevator.state.Floor][BtnHallUp] {
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

	return elevator, clearedOrders
}
