package localsingle

import (
	elevio "TTK4145/ElevIO"
)

func setCabLights(elevator LocalSingleElevator) {
	for floor := range N_FLOORS {
		elevio.SetButtonLamp(elevio.BT_Cab, floor, elevator.requests[floor][elevio.BT_Cab])
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
			// StartTimer(elevator.doorOpenDuration_s)
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
			// StartTimer()
			elevator.ClearAtCurrentFloor()
		case BehaviourMoving:
			elevio.SetMotorDirection(DirectionToMotorDirection(elevator.state.direction))
		case BehaviourIdle:

		}
	}
	setCabLights(*elevator)
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
			// StartTimer(doorOpenDuration_s)
			setCabLights(*elevator)
			elevator.state.behaviour = BehaviourDoorOpen
		}
	default:
		return
	}
}

func (elevator *LocalSingleElevator) FSM_OnDoorTimeout() {
	switch elevator.state.behaviour {
	case BehaviourDoorOpen:
		pair := elevator.ChooseDirection()
		elevator.state.direction = pair.direction
		elevator.state.behaviour = pair.behaviour

		switch elevator.state.behaviour {
		case BehaviourDoorOpen:
			// StartTimer(doorOpenDuration_s)
			elevator.ClearAtCurrentFloor()
			setCabLights(*elevator)
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
