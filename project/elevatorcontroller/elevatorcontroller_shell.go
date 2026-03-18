package elevatorcontroller

import (
	"fmt"
	elevio "project/elevio"
	"time"
)

func Run(
	assignedRequestsChan <-chan RequestMatrix,
	floorChan <-chan int,
	obstructionChan <-chan bool,

	driverCommandChan chan<- elevio.DriverCommand,
	clearedOrdersChan chan<- []Order,
	localStateChan chan<- ElevatorState,
) {
	fmt.Println("LocalSingleElevator started")
	elevator := makeUninitializedElevator()

	motorTimer := time.NewTimer(motorWatchdogTimeout)
	doorTimer := time.NewTimer(doorOpenDuration)
	doorTimer.Stop()
	motorTimer.Stop()

	var effects []effect

	if elevio.GetFloor() == -1 {
		elevator, effects = onInitBetweenFloors(elevator)
		applyEffects(effects, driverCommandChan, localStateChan, clearedOrdersChan, doorTimer, motorTimer)

	} else {
		elevator.state.Floor = elevio.GetFloor()
	}

	for {
		select {
		case newRequests := <-assignedRequestsChan:
			elevator, effects = onNewRequestMatrix(elevator, newRequests)
			applyEffects(effects, driverCommandChan, localStateChan, clearedOrdersChan, doorTimer, motorTimer)

		case floor := <-floorChan:
			motorTimer.Stop()
			elevator, effects = onFloorArrival(elevator, floor)
			applyEffects(effects, driverCommandChan, localStateChan, clearedOrdersChan, doorTimer, motorTimer)

		case obstructed := <-obstructionChan:
			elevator, effects = onObstruction(elevator, obstructed)
			applyEffects(effects, driverCommandChan, localStateChan, clearedOrdersChan, doorTimer, motorTimer)

		case <-doorTimer.C:
			elevator, effects = onDoorTimeout(elevator)
			applyEffects(effects, driverCommandChan, localStateChan, clearedOrdersChan, doorTimer, motorTimer)

		case <-motorTimer.C:
			elevator, effects = onMotorTimeout(elevator)
			applyEffects(effects, driverCommandChan, localStateChan, clearedOrdersChan, doorTimer, motorTimer)
		}
	}
}

func applyEffects(
	effects []effect,
	driverCommandChan chan<- elevio.DriverCommand,
	localStateChan chan<- ElevatorState,
	clearedChan chan<- []Order,
	doorTimer *time.Timer,
	motorTimer *time.Timer,
) {
	for _, effect := range effects {
		switch effect.kind {
		case setMotorDirection:
			dir := effect.value.(Direction)
			driverCommandChan <- elevio.DriverCommand{Kind: elevio.CommandSetMotorDirection, MotorDirection: directionToMotorDirection(dir)}
			if dir != DirStop {
				motorTimer.Reset(motorWatchdogTimeout)
			}
		case setDoorOpenLamp:
			value := effect.value.(bool)
			driverCommandChan <- elevio.DriverCommand{Kind: elevio.CommandSetDoorLamp, Value: value}
		case setFloorIndicator:
			floor := effect.value.(int)
			driverCommandChan <- elevio.DriverCommand{Kind: elevio.CommandSetFloorIndicator, Floor: floor}
		case setButtonLamp:
			args := effect.value.(buttonLampArgs)
			driverCommandChan <- elevio.DriverCommand{Kind: elevio.CommandSetButtonLamp, Button: ButtonTypeToElevio(args.Btn), Floor: args.Floor, Value: args.Value}
		case resetDoorTimer:
			doorTimer.Reset(doorOpenDuration)
		case publishClearedOrders:
			cleared := effect.value.([]Order)
			clearedChan <- cleared
		case publishLocalState:
			localStateChan <- effect.value.(ElevatorState)
		}

	}
}

func directionToMotorDirection(direction Direction) elevio.MotorDirection {
	switch direction {
	case DirUp:
		return elevio.MD_Up
	case DirDown:
		return elevio.MD_Down
	case DirStop:
		return elevio.MD_Stop
	default:
		return elevio.MD_Stop
	}
}

func ButtonTypeToElevio(b ButtonType) elevio.ButtonType {
	switch b {
	case BtnHallUp:
		return elevio.BT_HallUp
	case BtnHallDown:
		return elevio.BT_HallDown
	default:
		return elevio.BT_Cab
	}
}
