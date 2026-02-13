package localsingle

func (elevator *LocalSingleElevator) RequestsAbove() bool {
	for f := elevator.state.floor + 1; f < N_FLOORS; f++ {
		for btn := range N_BUTTONS {
			if elevator.requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func (elevator *LocalSingleElevator) RequestsBelow() bool {
	for f := range elevator.state.floor {
		for btn := range N_BUTTONS {
			if elevator.requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func (elevator *LocalSingleElevator) RequestsHere() bool {
	for btn := range N_BUTTONS {
		if elevator.requests[elevator.state.floor][btn] {
			return true
		}
	}
	return false
}

func (elevator *LocalSingleElevator) ChooseDirection() DirectionBehaviourPair {
	switch elevator.state.direction {
	case DirUp:
		if elevator.RequestsAbove() {
			return DirectionBehaviourPair{DirUp, BehaviourMoving}
		} else if elevator.RequestsHere() {
			return DirectionBehaviourPair{DirStop, BehaviourDoorOpen}
		} else if elevator.RequestsBelow() {
			return DirectionBehaviourPair{DirDown, BehaviourMoving}
		} else {
			return DirectionBehaviourPair{DirStop, BehaviourIdle}
		}
	case DirDown:
		if elevator.RequestsBelow() {
			return DirectionBehaviourPair{DirDown, BehaviourMoving}
		} else if elevator.RequestsHere() {
			return DirectionBehaviourPair{DirStop, BehaviourDoorOpen}
		} else if elevator.RequestsAbove() {
			return DirectionBehaviourPair{DirUp, BehaviourMoving}
		} else {
			return DirectionBehaviourPair{DirStop, BehaviourIdle}
		}
	case DirStop:
		if elevator.RequestsHere() {
			return DirectionBehaviourPair{DirStop, BehaviourDoorOpen}
		} else if elevator.RequestsAbove() {
			return DirectionBehaviourPair{DirUp, BehaviourMoving}
		} else if elevator.RequestsBelow() {
			return DirectionBehaviourPair{DirDown, BehaviourMoving}
		} else {
			return DirectionBehaviourPair{DirStop, BehaviourIdle}
		}
	default:
		return DirectionBehaviourPair{DirStop, BehaviourIdle}
	}
}

func (elevator *LocalSingleElevator) ShouldStop() bool {
	switch elevator.state.direction {
	case DirDown:
		return elevator.requests[elevator.state.floor][BtnHallDown] ||
			elevator.requests[elevator.state.floor][BtnCab] ||
			!elevator.RequestsBelow()
	case DirUp:
		return elevator.requests[elevator.state.floor][BtnHallUp] ||
			elevator.requests[elevator.state.floor][BtnCab] ||
			!elevator.RequestsAbove()
	default:
		return true
	}
}

func (elevator *LocalSingleElevator) ShouldClearImmediately(buttonFloor int, buttonType ButtonType) bool {
	return elevator.state.floor == buttonFloor &&
		((elevator.state.direction == DirUp && buttonType == BtnHallUp) ||
			(elevator.state.direction == DirDown && buttonType == BtnHallDown) ||
			elevator.state.direction == DirStop ||
			buttonType == BtnCab)
}


//Returnerer en liste over Ordre som ble fjernet
func (elevator *LocalSingleElevator) ClearAtCurrentFloor() []Order {
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
		
		if !elevator.RequestsAbove() && !elevator.requests[elevator.state.floor][BtnHallUp] && elevator.requests[elevator.state.floor][BtnHallDown] {
			elevator.requests[elevator.state.floor][BtnHallDown] = false
			clearedOrders = append(clearedOrders, Order{elevator.state.floor, BtnHallDown})
		}
		
	case DirDown:
		if elevator.requests[elevator.state.floor][BtnHallDown] {
			elevator.requests[elevator.state.floor][BtnHallDown] = false
			clearedOrders = append(clearedOrders, Order{elevator.state.floor, BtnHallDown})
		}

		if !elevator.RequestsBelow() && !elevator.requests[elevator.state.floor][BtnHallDown] && elevator.requests[elevator.state.floor][BtnHallUp] {
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