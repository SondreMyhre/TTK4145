package localsingle

import (
	elevio "Project/elevio"
	"fmt"
	"time"
)

const (
	doorOpenDuration = 3 * time.Second
)

// type Input struct {
// 	buttonChan <-chan elevio.ButtonEvent
// 	floorChan <-chan int
// 	obstructionChan <-chan bool
// }

// type Output struct {
// 	driverCommandChan chan<- elevio.DriverCommand
// 	clearedChan chan<- []Order
// 	localStateChan chan<- LocalSingleElevator
// }

func Run(
	localOrderChan <-chan elevio.ButtonEvent,
	floorChan <-chan int,
	obstructionChan <-chan bool,

	driverCommandChan chan<- elevio.DriverCommand,
	clearedOrdersChan chan<- []Order,
	localStateChan chan<- LocalSingleElevator, // Muligens kun sende state og ikke hele elevator
) {
	fmt.Println("LocalSingleElevator started")
	elevator := makeUninitializedElevator()

	doorTimer := time.NewTimer(doorOpenDuration)
	doorTimer.Stop()

	var commands []command

	if elevio.GetFloor() == -1 {
		commands = append(commands, elevator.onInitBetweenFloors())
		executeCommands(commands, driverCommandChan, clearedOrdersChan, doorTimer)
	}

	for {
		select {
		case buttonEvent := <-localOrderChan:
			btn := elevioToButtonType(buttonEvent.Button)
			commands = elevator.onRequestButtonPress(buttonEvent.Floor, btn)

		case floor := <-floorChan:
			commands = elevator.onFloorArrival(floor)

		case obstructed := <-obstructionChan:

			commands = elevator.onObstruction(obstructed)

		case <-doorTimer.C:
			commands = elevator.onDoorTimeout()

		}

		executeCommands(commands, driverCommandChan, clearedOrdersChan, doorTimer)
		localStateChan <- elevator // Sender state til OrderSync
	}
}

func executeCommands( // Kanskje det er rotete å ha den slik når det ikke kjøres som en egen goroutine, og heller eksplisitt execute commands i hver case i Run()?
	commands []command,
	driverCommandChan chan<- elevio.DriverCommand,
	clearedChan chan<- []Order,
	doorTimer *time.Timer,
) {
	for _, command := range commands {
		switch command._type {
		case setMotorDirection:
			dir := command.value.(direction)
			driverCommandChan <- elevio.DriverCommand{Type: elevio.CommandSetMotorDirection, MotorDirection: directionToMotorDirection(dir)}
		case setDoorOpenLamp:
			value := command.value.(bool)
			driverCommandChan <- elevio.DriverCommand{Type: elevio.CommandSetDoorLamp, Value: value}
		case setButtonLamp:
			args := command.value.(buttonLampArgs)
			driverCommandChan <- elevio.DriverCommand{Type: elevio.CommandSetButtonLamp, Button: buttonTypeToElevio(args.Btn), Floor: args.Floor, Value: args.Value}
		case resetDoorTimer:
			doorTimer.Reset(doorOpenDuration)
		case sendClearedOrders:
			cleared := command.value.([]Order)
			clearedChan <- cleared
		}

	}
}

func directionToMotorDirection(direction direction) elevio.MotorDirection {
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

func buttonTypeToElevio(b buttonType) elevio.ButtonType {
	switch b {
	case BtnHallUp:
		return elevio.BT_HallUp
	case BtnHallDown:
		return elevio.BT_HallDown
	default:
		return elevio.BT_Cab
	}
}

func elevioToButtonType(b elevio.ButtonType) buttonType {
	switch b {
	case elevio.BT_HallUp:
		return BtnHallUp
	case elevio.BT_HallDown:
		return BtnHallDown
	default:
		return BtnCab
	}
}
