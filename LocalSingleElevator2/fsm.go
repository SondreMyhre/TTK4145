package localsingle2

import (
	elevio "TTK4145/ElevIO"
	"fmt"
)

// FSM encapsulates the finite state machine for a single elevator.
type FSM struct {
	elevator Elevator
}

// NewFSM creates a new FSM.
func NewFSM() *FSM {
	return &FSM{
		elevator: NewElevator(),
	}
}

// GetElevator returns a copy of the current elevator state.
func (fsm *FSM) GetElevator() Elevator {
	return fsm.elevator
}

// SetDoorOpenDuration configures the door open duration.
func (fsm *FSM) SetDoorOpenDuration(duration float64) {
	fsm.elevator.Config.DoorOpenDuration = duration
}

// setAllLights updates all request button lights based on current state.
func (fsm *FSM) setAllLights() {
	for floor := 0; floor < NumFloors; floor++ {
		for btn := 0; btn < NumButtons; btn++ {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, fsm.elevator.Requests[floor][btn])
		}
	}
}

// OnInitBetweenFloors handles initialization when the elevator starts between floors.
func (fsm *FSM) OnInitBetweenFloors() {
	elevio.SetMotorDirection(elevio.MotorDirection(DirDown))
	fsm.elevator.Direction = DirDown
	fsm.elevator.Behaviour = BehaviourMoving
}

// OnRequestButtonPress handles a button press event.
func (fsm *FSM) OnRequestButtonPress(btnFloor int, btnType ButtonType) {
	fmt.Printf("\n\nOnRequestButtonPress(%d, %s)\n", btnFloor, btnType)
	fsm.elevator.Print()

	switch fsm.elevator.Behaviour {
	case BehaviourDoorOpen:
		if ShouldClearImmediately(fsm.elevator, btnFloor, btnType) {
			TimerStart(fsm.elevator.Config.DoorOpenDuration)
		} else {
			fsm.elevator.Requests[btnFloor][btnType] = true
		}

	case BehaviourMoving:
		fsm.elevator.Requests[btnFloor][btnType] = true

	case BehaviourIdle:
		fsm.elevator.Requests[btnFloor][btnType] = true
		pair := ChooseDirection(fsm.elevator)
		fsm.elevator.Direction = pair.Direction
		fsm.elevator.Behaviour = pair.Behaviour

		switch pair.Behaviour {
		case BehaviourDoorOpen:
			elevio.SetDoorOpenLamp(true)
			TimerStart(fsm.elevator.Config.DoorOpenDuration)
			fsm.elevator = ClearAtCurrentFloor(fsm.elevator)

		case BehaviourMoving:
			elevio.SetMotorDirection(elevio.MotorDirection(fsm.elevator.Direction))

		case BehaviourIdle:
			// Nothing to do.
		}
	}

	fsm.setAllLights()

	fmt.Println("\nNew state:")
	fsm.elevator.Print()
}

// OnFloorArrival handles arrival at a floor.
func (fsm *FSM) OnFloorArrival(newFloor int) {
	fmt.Printf("\n\nOnFloorArrival(%d)\n", newFloor)
	fsm.elevator.Print()

	fsm.elevator.Floor = newFloor
	elevio.SetFloorIndicator(fsm.elevator.Floor)

	switch fsm.elevator.Behaviour {
	case BehaviourMoving:
		if ShouldStop(fsm.elevator) {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			fsm.elevator = ClearAtCurrentFloor(fsm.elevator)
			TimerStart(fsm.elevator.Config.DoorOpenDuration)
			fsm.setAllLights()
			fsm.elevator.Behaviour = BehaviourDoorOpen
		}

	default:
		// Nothing to do.
	}

	fmt.Println("\nNew state:")
	fsm.elevator.Print()
}

// OnDoorTimeout handles the door timeout event.
func (fsm *FSM) OnDoorTimeout() {
	fmt.Println("\n\nOnDoorTimeout()")
	fsm.elevator.Print()

	switch fsm.elevator.Behaviour {
	case BehaviourDoorOpen:
		pair := ChooseDirection(fsm.elevator)
		fsm.elevator.Direction = pair.Direction
		fsm.elevator.Behaviour = pair.Behaviour

		switch fsm.elevator.Behaviour {
		case BehaviourDoorOpen:
			TimerStart(fsm.elevator.Config.DoorOpenDuration)
			fsm.elevator = ClearAtCurrentFloor(fsm.elevator)
			fsm.setAllLights()

		case BehaviourMoving, BehaviourIdle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevio.MotorDirection(fsm.elevator.Direction))
		}

	default:
		// Nothing to do.
	}

	fmt.Println("\nNew state:")
	fsm.elevator.Print()
}
