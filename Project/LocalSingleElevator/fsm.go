package localsingle

import (
	elevio "TTK4145/ElevIO"
)

func setAllLights(elevator LocalSingleElevator) {	// Bør endres til setCabLights hvis kun Cab eies av localsingle
	for floor := range N_FLOORS {
		for btn := range N_BUTTONS {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, elevator.requests[floor][btn])
		}
	}
}

func (elevator *LocalSingleElevator) FSM_OnInitBetweenFloors() {
	elevio.SetMotorDirection(elevio.MD_Down)
	elevator.state.direction = DirDown
	elevator.state.behaviour = BehaviourMoving
}

func (elevator *LocalSingleElevator) FSM_OnRequestButtonPress(buttonFloor int, buttonType elevio.ButtonType) {
	switch elevator.state.behaviour {
	case BehaviourDoorOpen:
		if elevator.ShouldClearImmediately(buttonFloor, buttonType) {
			elevator.ResetDoorTimer()
		} else {
			elevator.requests[buttonFloor][buttonType] = true
		}
	case BehaviourMoving:
		elevator.requests[buttonFloor][buttonType] = true
	case BehaviourIdle:
		elevator.requests[buttonFloor][buttonType] = true
		pair := elevator.ChooseDirection()
		elevator.state.direction = pair.direction
		elevator.state.behaviour = pair.behaviour
		switch pair.behaviour {
		case BehaviourDoorOpen:
			elevio.SetDoorOpenLamp(true)
			elevator.ResetDoorTimer()
			elevator.ClearAtCurrentFloor()
		case BehaviourMoving:
			elevio.SetMotorDirection(DirectionToMotorDirection(elevator.state.direction))
		case BehaviourIdle:

		}
	}
	setAllLights(*elevator)
}

func (elevator *LocalSingleElevator) FSM_OnFloorArrival(newFloor int) {
	elevator.state.floor = newFloor
	elevio.SetFloorIndicator(elevator.state.floor)

	switch elevator.state.behaviour {
	case BehaviourMoving:
		if elevator.ShouldStop() {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			elevator.ClearAtCurrentFloor()
			elevator.ResetDoorTimer()
			setAllLights(*elevator)
			elevator.state.behaviour = BehaviourDoorOpen
		}
	default:
		return
	}
}

func (elevator *LocalSingleElevator) FSM_OnDoorTimeout() {
	elevator.ResetDoorTimer()
	switch elevator.state.behaviour {
	case BehaviourDoorOpen:
		pair := elevator.ChooseDirection()
		elevator.state.direction = pair.direction
		elevator.state.behaviour = pair.behaviour

		switch elevator.state.behaviour {
		case BehaviourDoorOpen:
			elevator.ResetDoorTimer()
			elevator.ClearAtCurrentFloor()
			setAllLights(*elevator)
		case BehaviourMoving:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(DirectionToMotorDirection(elevator.state.direction))
		case BehaviourIdle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(DirectionToMotorDirection(elevator.state.direction))
		}
	default:
		return
	}
}
