package localsingle

import (
	elevio "Project/ElevIO"
)

type DirectionBehaviourPair struct {
	direction Direction
	behaviour ElevatorBehaviour
}

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
		return elevator.requests[elevator.state.floor][elevio.BT_HallDown] ||
			elevator.requests[elevator.state.floor][elevio.BT_Cab] ||
			!elevator.RequestsBelow()
	case DirUp:
		return elevator.requests[elevator.state.floor][elevio.BT_HallUp] ||
			elevator.requests[elevator.state.floor][elevio.BT_Cab] ||
			!elevator.RequestsAbove()
	default:
		return true
	}
}

func (elevator *LocalSingleElevator) ShouldClearImmediately(buttonFloor int, buttonType elevio.ButtonType) bool {
	return elevator.state.floor == buttonFloor &&
		((elevator.state.direction == DirUp && buttonType == elevio.BT_HallUp) ||
			(elevator.state.direction == DirDown && buttonType == elevio.BT_HallDown) ||
			elevator.state.direction == DirStop ||
			buttonType == elevio.BT_Cab)
}

func (elevator *LocalSingleElevator) ClearAtCurrentFloor() {
	elevator.requests[elevator.state.floor][elevio.BT_Cab] = false
	switch elevator.state.direction {
	case DirUp:
		if !elevator.RequestsAbove() && !elevator.requests[elevator.state.floor][elevio.BT_HallUp] {
			elevator.requests[elevator.state.floor][elevio.BT_HallDown] = false
		}
		elevator.requests[elevator.state.floor][elevio.BT_HallUp] = false
	case DirDown:
		if !elevator.RequestsBelow() && !elevator.requests[elevator.state.floor][elevio.BT_HallDown] {
			elevator.requests[elevator.state.floor][elevio.BT_HallUp] = false
		}
		elevator.requests[elevator.state.floor][elevio.BT_HallDown] = false
	default:
		elevator.requests[elevator.state.floor][elevio.BT_HallUp] = false
		elevator.requests[elevator.state.floor][elevio.BT_HallDown] = false
	}
}
