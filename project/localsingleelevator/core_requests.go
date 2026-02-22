package localsingle

func (elevator *LocalSingleElevator) requestsAbove() bool {
	for f := elevator.State.Floor + 1; f < N_FLOORS; f++ {
		for btn := range N_BUTTONS {
			if elevator.requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func (elevator *LocalSingleElevator) requestsBelow() bool {
	for f := range elevator.State.Floor {
		for btn := range N_BUTTONS {
			if elevator.requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func (elevator *LocalSingleElevator) requestsHere() bool {
	for btn := range N_BUTTONS {
		if elevator.requests[elevator.State.Floor][btn] {
			return true
		}
	}
	return false
}

func (elevator *LocalSingleElevator) chooseDirection() directionBehaviourPair {
	switch elevator.State.Direction {
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

func (elevator *LocalSingleElevator) shouldStop() bool {
	switch elevator.State.Direction {
	case DirDown:
		return elevator.requests[elevator.State.Floor][BtnHallDown] ||
			elevator.requests[elevator.State.Floor][BtnCab] ||
			!elevator.requestsBelow()
	case DirUp:
		return elevator.requests[elevator.State.Floor][BtnHallUp] ||
			elevator.requests[elevator.State.Floor][BtnCab] ||
			!elevator.requestsAbove()
	default:
		return true
	}
}

func (elevator *LocalSingleElevator) shouldClearImmediately(buttonFloor int, buttonType ButtonType) bool {
	return elevator.State.Floor == buttonFloor &&
		((elevator.State.Direction == DirUp && buttonType == BtnHallUp) ||
			(elevator.State.Direction == DirDown && buttonType == BtnHallDown) ||
			elevator.State.Direction == DirStop ||
			buttonType == BtnCab)
}

// Returnerer en liste over Ordre som ble fjernet
func (elevator *LocalSingleElevator) clearAtCurrentFloor() []Order {
	var clearedOrders []Order

	if elevator.requests[elevator.State.Floor][BtnCab] {
		elevator.requests[elevator.State.Floor][BtnCab] = false
		clearedOrders = append(clearedOrders, Order{elevator.State.Floor, BtnCab})
	}

	switch elevator.State.Direction {
	case DirUp:
		if elevator.requests[elevator.State.Floor][BtnHallUp] {
			elevator.requests[elevator.State.Floor][BtnHallUp] = false
			clearedOrders = append(clearedOrders, Order{elevator.State.Floor, BtnHallUp})
		}

		if !elevator.requestsAbove() && !elevator.requests[elevator.State.Floor][BtnHallUp] && elevator.requests[elevator.State.Floor][BtnHallDown] {
			elevator.requests[elevator.State.Floor][BtnHallDown] = false
			clearedOrders = append(clearedOrders, Order{elevator.State.Floor, BtnHallDown})
		}

	case DirDown:
		if elevator.requests[elevator.State.Floor][BtnHallDown] {
			elevator.requests[elevator.State.Floor][BtnHallDown] = false
			clearedOrders = append(clearedOrders, Order{elevator.State.Floor, BtnHallDown})
		}

		if !elevator.requestsBelow() && !elevator.requests[elevator.State.Floor][BtnHallDown] && elevator.requests[elevator.State.Floor][BtnHallUp] {
			elevator.requests[elevator.State.Floor][BtnHallUp] = false
			clearedOrders = append(clearedOrders, Order{elevator.State.Floor, BtnHallUp})
		}

	default:
		if elevator.requests[elevator.State.Floor][BtnHallUp] {
			elevator.requests[elevator.State.Floor][BtnHallUp] = false
			clearedOrders = append(clearedOrders, Order{elevator.State.Floor, BtnHallUp})
		}

		if elevator.requests[elevator.State.Floor][BtnHallDown] {
			elevator.requests[elevator.State.Floor][BtnHallDown] = false
			clearedOrders = append(clearedOrders, Order{elevator.State.Floor, BtnHallDown})
		}
	}

	return clearedOrders
}

func (elevator *LocalSingleElevator) generateLightCommands() []command {
	commands := make([]command, 0, N_FLOORS*N_BUTTONS)
	for f := range N_FLOORS {
		for btn := range N_BUTTONS {
			commands = append(commands, command{_type: setButtonLamp, value: buttonLampArgs{f, ButtonType(btn), elevator.requests[f][btn]}})
		}
	}

	return commands
}
