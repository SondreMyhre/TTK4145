package localsingle

import (
	"context"
	"fmt"
	elevio "project/elevio"
	"time"
)

const (
	doorOpenDuration = 3 * time.Second
	motorWatchdogTimeout = 5 * time.Second
)

func Run(
	ctx context.Context,
	requestMatrixChan <-chan RequestMatrix,
	floorChan <-chan int,
	obstructionChan <-chan bool,

	driverCommandChan chan<- elevio.DriverCommand,
	clearedOrdersChan chan<- []Order,
	localStateChan chan<- ElevatorState,
) error {
	fmt.Println("LocalSingleElevator started")
	elevator := makeUninitializedElevator()

	doorTimer := time.NewTimer(doorOpenDuration)
	doorTimer.Stop()

	motorWatchdogTimer := time.NewTimer(motorWatchdogTimeout)
	motorWatchdogTimer.Stop()

	var commands []command



	if elevio.GetFloor() == -1 {
		commands = append(commands, elevator.onInitBetweenFloors()...)
		executeCommands(commands, driverCommandChan, localStateChan, clearedOrdersChan, doorTimer, motorWatchdogTimer)
	}

	for {
		select {
		case <-ctx.Done():
			return nil

		case newRequests := <-requestMatrixChan:
			commands = elevator.onNewRequestMatrix(newRequests)
			executeCommands(commands, driverCommandChan, localStateChan, clearedOrdersChan, doorTimer, motorWatchdogTimer)

		case floor := <-floorChan:
			motorWatchdogTimer.Stop()
			commands = elevator.onFloorArrival(floor)
			executeCommands(commands, driverCommandChan, localStateChan, clearedOrdersChan, doorTimer, motorWatchdogTimer)

		case obstructed := <-obstructionChan:
			commands = elevator.onObstruction(obstructed)
			executeCommands(commands, driverCommandChan, localStateChan, clearedOrdersChan, doorTimer, motorWatchdogTimer)

		case <-doorTimer.C:
			commands = elevator.onDoorTimeout()
			executeCommands(commands, driverCommandChan, localStateChan, clearedOrdersChan, doorTimer, motorWatchdogTimer)

		case <-motorWatchdogTimer.C:
			commands = elevator.onMotorTimeout()
			executeCommands(commands, driverCommandChan, localStateChan, clearedOrdersChan, doorTimer, motorWatchdogTimer)
		}
	}
}

func executeCommands(
	commands []command,
	driverCommandChan chan<- elevio.DriverCommand,
	localStateChan chan<- ElevatorState,
	clearedChan chan<- []Order,
	doorTimer *time.Timer,
	motorWatchdogTimer *time.Timer,
) {
	for _, command := range commands {
		switch command._type {
		case setMotorDirection:
			dir := command.value.(Direction)
			driverCommandChan <- elevio.DriverCommand{Type: elevio.CommandSetMotorDirection, MotorDirection: directionToMotorDirection(dir)}
			if dir != DirStop {
				motorWatchdogTimer.Reset(motorWatchdogTimeout)
			}
		case setDoorOpenLamp:
			value := command.value.(bool)
			driverCommandChan <- elevio.DriverCommand{Type: elevio.CommandSetDoorLamp, Value: value}
		case setFloorIndicator:
			floor := command.value.(int)
			driverCommandChan <- elevio.DriverCommand{Type: elevio.CommandSetFloorIndicator, Floor: floor}
		case setButtonLamp:
			args := command.value.(buttonLampArgs)
			driverCommandChan <- elevio.DriverCommand{Type: elevio.CommandSetButtonLamp, Button: ButtonTypeToElevio(args.Btn), Floor: args.Floor, Value: args.Value}
		case resetDoorTimer:
			doorTimer.Reset(doorOpenDuration)
		case sendClearedOrders:
			cleared := command.value.([]Order)
			clearedChan <- cleared
		case sendLocalState:
			localStateChan <- command.value.(ElevatorState)
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

func elevioToButtonType(b elevio.ButtonType) ButtonType {
	switch b {
	case elevio.BT_HallUp:
		return BtnHallUp
	case elevio.BT_HallDown:
		return BtnHallDown
	default:
		return BtnCab
	}
}
