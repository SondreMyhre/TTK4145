package localsingle

import (
    elevio "TTK4145/ElevIO"
)

type DirectionBehaviourPair struct {
    direction Direction
    behaviour ElevatorBehaviour
}

func (e *LocalSingleElevator) requestsAbove() bool {
    for f := e.state.floor+1; f < N_FLOORS; f++ {
        for btn := range N_BUTTONS {
            if (e.requests[f][btn]){
                return true
            }
        }
    }
    return false
}

func (e *LocalSingleElevator) requestsBelow() bool {
    for f := range e.state.floor {
        for btn := range N_BUTTONS {
            if (e.requests[f][btn]){
                return true
            }
        }
    }
    return false
}

func (e *LocalSingleElevator) requestsHere() bool {
    for btn := range N_BUTTONS {
        if (e.requests[e.state.floor][btn]){
            return true
        }
    }
    return false
}

func (e *LocalSingleElevator) ChooseDirection() DirectionBehaviourPair {
    switch e.state.direction {
    case DirUp:
        if e.requestsAbove() {
            return DirectionBehaviourPair{DirUp, BehaviourMoving}
        } else if e.requestsHere() {
            return DirectionBehaviourPair{DirStop, BehaviourDoorOpen}
        } else if e.requestsBelow() {
            return DirectionBehaviourPair{DirDown, BehaviourMoving}
        } else {
            return DirectionBehaviourPair{DirStop, BehaviourIdle}
        }
    case DirDown:
        if e.requestsBelow() {
            return DirectionBehaviourPair{DirDown, BehaviourMoving}
        } else if e.requestsHere() {
            return DirectionBehaviourPair{DirStop, BehaviourDoorOpen}
        } else if e.requestsAbove() {
            return DirectionBehaviourPair{DirUp, BehaviourMoving}
        } else {
            return DirectionBehaviourPair{DirStop, BehaviourIdle}
        }
    case DirStop:
        if e.requestsHere() {
            return DirectionBehaviourPair{DirStop, BehaviourIdle}
        } else if e.requestsAbove() {
            return DirectionBehaviourPair{DirUp, BehaviourMoving}
        } else if e.requestsBelow() {
            return DirectionBehaviourPair{DirDown, BehaviourMoving}
        } else {
            return DirectionBehaviourPair{DirStop, BehaviourIdle}
        }
    default:
        return DirectionBehaviourPair{DirStop, BehaviourIdle}
    }
}


func (e *LocalSingleElevator) ShouldStop() bool {
    switch e.state.direction {
    case DirDown: 
        return e.requests[e.state.floor][elevio.BT_HallDown] ||
               e.requests[e.state.floor][elevio.BT_Cab]      ||
               !e.requestsBelow()
    case DirUp: 
        return e.requests[e.state.floor][elevio.BT_HallUp] || 
               e.requests[e.state.floor][elevio.BT_Cab]    || 
               !e.requestsAbove()
    default:
        return true
    }
}

func (e *LocalSingleElevator) ShouldClearImmediately(buttonFloor int, buttonType elevio.ButtonType) bool {
    switch e.state.direction {
    case DirDown: 
        return e.requests[e.state.floor][elevio.BT_HallDown] ||
               e.requests[e.state.floor][elevio.BT_Cab]      ||
               !e.requestsBelow()
    case DirUp: 
        return e.requests[e.state.floor][elevio.BT_HallUp] || 
               e.requests[e.state.floor][elevio.BT_Cab]    || 
               !e.requestsAbove()
    default:
        return true
    }
}

func (e *LocalSingleElevator) shouldClearImmediately(buttonFloor int, buttonType elevio.ButtonType) bool {
    return e.state.floor == buttonFloor && 
           ((e.state.direction == DirUp && buttonType == elevio.BT_HallUp)     ||
            (e.state.direction == DirDown && buttonType == elevio.BT_HallDown) ||
            e.state.direction == DirStop ||
            buttonType == elevio.BT_Cab)
}

func (e *LocalSingleElevator) ClearAtCurrentFloor() {
		e.requests[e.state.floor][elevio.BT_Cab] = false
		switch e.state.direction {
		case DirUp:
			if !e.requestsAbove() && !e.requests[e.state.floor][elevio.BT_HallUp] {
				e.requests[e.state.floor][elevio.BT_HallDown] = false
			}
			e.requests[e.state.floor][elevio.BT_HallUp] = false
		case DirDown:
			if !e.requestsBelow() && !e.requests[e.state.floor][elevio.BT_HallDown] {
				e.requests[e.state.floor][elevio.BT_HallUp] = false
			}
			e.requests[e.state.floor][elevio.BT_HallDown] = false
		default:
			e.requests[e.state.floor][elevio.BT_HallUp] = false
			e.requests[e.state.floor][elevio.BT_HallDown] = false
		}
}
