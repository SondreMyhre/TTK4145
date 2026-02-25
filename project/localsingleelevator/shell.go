package localsingle

import (
	"fmt"
	elevio "project/elevio"
	"time"
	"context"
)

const (
	doorOpenDuration = 3 * time.Second
)

func Run(
	ctx context.Context,
	localOrderChan <-chan elevio.ButtonEvent,
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

	localStateTicker := time.NewTicker(100 * time.Millisecond)

	var commands []command

	if elevio.GetFloor() == -1 {
		commands = append(commands, elevator.onInitBetweenFloors())
		executeCommands(commands, driverCommandChan, clearedOrdersChan, doorTimer)
	}

	for {
		select {
		case <- ctx.Done():
			return nil
		case buttonEvent := <-localOrderChan:
			btn := elevioToButtonType(buttonEvent.Button)
			commands = elevator.onRequestButtonPress(buttonEvent.Floor, btn)
			executeCommands(commands, driverCommandChan, clearedOrdersChan, doorTimer)

		case floor := <-floorChan:
			commands = elevator.onFloorArrival(floor)
			executeCommands(commands, driverCommandChan, clearedOrdersChan, doorTimer)

		case obstructed := <-obstructionChan:
			commands = elevator.onObstruction(obstructed)
			executeCommands(commands, driverCommandChan, clearedOrdersChan, doorTimer)

		case <-doorTimer.C:
			commands = elevator.onDoorTimeout()
			executeCommands(commands, driverCommandChan, clearedOrdersChan, doorTimer)

		case <-localStateTicker.C:
			select {
			case localStateChan <- elevator.State:
			default:
			}
		}
	}
}

func executeCommands(
	commands []command,
	driverCommandChan chan<- elevio.DriverCommand,
	clearedChan chan<- []Order,
	doorTimer *time.Timer,
) {
	for _, command := range commands {
		switch command._type {
		case setMotorDirection:
			dir := command.value.(Direction)
			driverCommandChan <- elevio.DriverCommand{Type: elevio.CommandSetMotorDirection, MotorDirection: directionToMotorDirection(dir)}
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
