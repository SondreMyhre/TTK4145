package localsingle

import (
	elevio "Project/ElevIO"
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
// 	driverCmdChan chan<- elevio.DriverCmd
// 	clearedChan chan<- []Order
// 	stateOutChan chan<- LocalSingleElevator
// }

func Run(
	buttonChan <-chan elevio.ButtonEvent,
	floorChan <-chan int,
	obstructionChan <-chan bool,

	driverCmdChan chan<- elevio.DriverCmd,
	clearedOrdersChan chan<- []Order,
	stateOutChan chan<- LocalSingleElevator, // Muligens kun sende state og ikke hele elevator
) {
	fmt.Println("LocalSingleElevator started")
	elevator := makeUninitializedElevator()

	doorTimer := time.NewTimer(doorOpenDuration)
	doorTimer.Stop()

	var commands []Command

	if elevio.GetFloor() == -1 {
		commands = append(commands, elevator.onInitBetweenFloors())
		executeCommands(commands, driverCmdChan, clearedOrdersChan, doorTimer)
	}

	for {
		select {
		case buttonEvent := <-buttonChan:
			btn := elevioToButtonType(buttonEvent.Button)
			commands = elevator.onRequestButtonPress(buttonEvent.Floor, btn)

		case floor := <-floorChan:
			commands = elevator.onFloorArrival(floor)

		case obstructed := <-obstructionChan:

			commands = elevator.onObstruction(obstructed)

		case <-doorTimer.C:
			commands = elevator.onDoorTimeout()

		}

		executeCommands(commands, driverCmdChan, clearedOrdersChan, doorTimer)
		stateOutChan <- elevator // Sender state til OrderSync
	}
}

func executeCommands( // Kanskje det er rotete å ha den slik når det ikke kjøres som en egen goroutine, og heller eksplisitt execute commands i hver case i Run()?
	commands []Command,
	driverCmdChan chan<- elevio.DriverCmd,
	clearedChan chan<- []Order,
	doorTimer *time.Timer,
) {
	for _, command := range commands {
		switch command._type {
		case setMotorDirection:
			dir := command.value.(Direction)
			driverCmdChan <- elevio.DriverCmd{Type: elevio.CmdSetMotorDirection, MotorDirection: directionToMotorDirection(dir)}
		case setDoorOpenLamp:
			value := command.value.(bool)
			driverCmdChan <- elevio.DriverCmd{Type: elevio.CmdSetDoorLamp, Value: value}
		case setButtonLamp:
			args := command.value.(ButtonLampArgs)
			driverCmdChan <- elevio.DriverCmd{Type: elevio.CmdSetButtonLamp, Button: buttonTypeToElevio(args.Btn), Floor: args.Floor, Value: args.Value}
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

func buttonTypeToElevio(b ButtonType) elevio.ButtonType {
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
