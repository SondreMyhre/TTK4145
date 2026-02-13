package localsingle

func (elevator *LocalSingleElevator) requestsAbove() bool {
	for f := elevator.state.floor + 1; f < N_FLOORS; f++ {
		for btn := range N_BUTTONS {
			if elevator.requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func (elevator *LocalSingleElevator) requestsBelow() bool {
	for f := range elevator.state.floor {
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
		if elevator.requests[elevator.state.floor][btn] {
			return true
		}
	}
	return false
}

func (elevator *LocalSingleElevator) chooseDirection() DirectionBehaviourPair {
	switch elevator.state.direction {
	case DirUp:
		if elevator.requestsAbove() {
			return DirectionBehaviourPair{DirUp, BehaviourMoving}
		} else if elevator.requestsHere() {
			return DirectionBehaviourPair{DirStop, BehaviourDoorOpen}
		} else if elevator.requestsBelow() {
			return DirectionBehaviourPair{DirDown, BehaviourMoving}
		} else {
			return DirectionBehaviourPair{DirStop, BehaviourIdle}
		}
	case DirDown:
		if elevator.requestsBelow() {
			return DirectionBehaviourPair{DirDown, BehaviourMoving}
		} else if elevator.requestsHere() {
			return DirectionBehaviourPair{DirStop, BehaviourDoorOpen}
		} else if elevator.requestsAbove() {
			return DirectionBehaviourPair{DirUp, BehaviourMoving}
		} else {
			return DirectionBehaviourPair{DirStop, BehaviourIdle}
		}
	case DirStop:
		if elevator.requestsHere() {
			return DirectionBehaviourPair{DirStop, BehaviourDoorOpen}
		} else if elevator.requestsAbove() {
			return DirectionBehaviourPair{DirUp, BehaviourMoving}
		} else if elevator.requestsBelow() {
			return DirectionBehaviourPair{DirDown, BehaviourMoving}
		} else {
			return DirectionBehaviourPair{DirStop, BehaviourIdle}
		}
	default:
		return DirectionBehaviourPair{DirStop, BehaviourIdle}
	}
}

func (elevator *LocalSingleElevator) shouldStop() bool {
	switch elevator.state.direction {
	case DirDown:
		return elevator.requests[elevator.state.floor][BtnHallDown] ||
			elevator.requests[elevator.state.floor][BtnCab] ||
			!elevator.requestsBelow()
	case DirUp:
		return elevator.requests[elevator.state.floor][BtnHallUp] ||
			elevator.requests[elevator.state.floor][BtnCab] ||
			!elevator.requestsAbove()
	default:
		return true
	}
}

func (elevator *LocalSingleElevator) shouldClearImmediately(buttonFloor int, buttonType ButtonType) bool {
	return elevator.state.floor == buttonFloor &&
		((elevator.state.direction == DirUp && buttonType == BtnHallUp) ||
			(elevator.state.direction == DirDown && buttonType == BtnHallDown) ||
			elevator.state.direction == DirStop ||
			buttonType == BtnCab)
}

// Returnerer en liste over Ordre som ble fjernet
func (elevator *LocalSingleElevator) clearAtCurrentFloor() []Order {
	var clearedOrders []Order

	if elevator.requests[elevator.state.floor][BtnCab] {
		elevator.requests[elevator.state.floor][BtnCab] = false
		clearedOrders = append(clearedOrders, Order{elevator.state.floor, BtnCab})
	}

	switch elevator.state.direction {
	case DirUp:
		if elevator.requests[elevator.state.floor][BtnHallUp] {
			elevator.requests[elevator.state.floor][BtnHallUp] = false
			clearedOrders = append(clearedOrders, Order{elevator.state.floor, BtnHallUp})
		}

		if !elevator.requestsAbove() && !elevator.requests[elevator.state.floor][BtnHallUp] && elevator.requests[elevator.state.floor][BtnHallDown] {
			elevator.requests[elevator.state.floor][BtnHallDown] = false
			clearedOrders = append(clearedOrders, Order{elevator.state.floor, BtnHallDown})
		}

	case DirDown:
		if elevator.requests[elevator.state.floor][BtnHallDown] {
			elevator.requests[elevator.state.floor][BtnHallDown] = false
			clearedOrders = append(clearedOrders, Order{elevator.state.floor, BtnHallDown})
		}

		if !elevator.requestsBelow() && !elevator.requests[elevator.state.floor][BtnHallDown] && elevator.requests[elevator.state.floor][BtnHallUp] {
			elevator.requests[elevator.state.floor][BtnHallUp] = false
			clearedOrders = append(clearedOrders, Order{elevator.state.floor, BtnHallUp})
		}

	default:
		if elevator.requests[elevator.state.floor][BtnHallUp] {
			elevator.requests[elevator.state.floor][BtnHallUp] = false
			clearedOrders = append(clearedOrders, Order{elevator.state.floor, BtnHallUp})
		}

		if elevator.requests[elevator.state.floor][BtnHallDown] {
			elevator.requests[elevator.state.floor][BtnHallDown] = false
			clearedOrders = append(clearedOrders, Order{elevator.state.floor, BtnHallDown})
		}
	}

	return clearedOrders
}

func (elevator *LocalSingleElevator) generateLightCommands() []Command {
	commands := make([]Command, 0, N_FLOORS*N_BUTTONS)
	for f := range N_FLOORS {
		for btn := range N_BUTTONS {
			commands = append(commands, Command{_type: setButtonLamp, value: ButtonLampArgs{f, ButtonType(btn), elevator.requests[f][btn]}})
		}
	}

	return commands
}
