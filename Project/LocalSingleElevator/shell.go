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
// 	buttonCh <-chan elevio.ButtonEvent
// 	osv. Samme for output
// }

func Run(
	buttonCh <-chan elevio.ButtonEvent,
	floorCh <-chan int,
	obstructionCh <-chan bool, 
	
	driverCmdCh chan<- elevio.DriverCmd,
	clearedCh chan<- []Order,
	stateOutCh chan <- LocalSingleElevator,
) {
	fmt.Println("LocalSingleElevator started")
	elevator := MakeUninitializedElevator()

	doorTimer := time.NewTimer(doorOpenDuration)
    doorTimer.Stop()

	var commands []Command

	if elevio.GetFloor() == -1 {
		commands = append(commands, elevator.FSM_OnInitBetweenFloors())
		executeCommands(commands, driverCmdCh, clearedCh, doorTimer)
	}

	for {
		select {
		case evt := <-buttonCh:
			btn := elevioToButtonType(evt.Button)
			commands = elevator.FSM_OnRequestButtonPress(evt.Floor, btn)

		case floor := <-floorCh:
			commands = elevator.FSM_OnFloorArrival(floor)
		
		case obstructed := <-obstructionCh:
			
			commands = elevator.onObstruction(obstructed)			

		case <-doorTimer.C:
            commands = elevator.FSM_OnDoorTimeout()
            
        // default:
		// 	time.Sleep(10 * time.Millisecond)	// Muligens vi ikke trenger
		}

		executeCommands(commands, driverCmdCh, clearedCh, doorTimer)
		stateOutCh <- elevator	// Sender state til OrderSync
	}
}

func executeCommands(
	commands []Command,
	driverCmdCh chan<- elevio.DriverCmd,
	clearedCh chan<- []Order,
	doorTimer *time.Timer,
) {
	for _, command := range commands {
		switch command._type {
		case setMotorDirection:
			dir := command.value.(Direction)
			driverCmdCh <- elevio.DriverCmd{Type: elevio.CmdSetMotorDirection, MotorDirection: directionToMotorDirection(dir)}
		case setDoorOpenLamp:
			value := command.value.(bool)
			driverCmdCh <- elevio.DriverCmd{Type: elevio.CmdSetDoorLamp, Value: value}
		case setButtonLamp:
			args := command.value.(ButtonLampArgs)
			driverCmdCh <- elevio.DriverCmd{Type: elevio.CmdSetButtonLamp, Button: buttonTypeToElevio(args.Btn), Floor: args.Floor, Value: args.Value}
		case ResetDoorTimer:
			doorTimer.Reset(doorOpenDuration)
		case sendClearedOrders:
			cleared := command.value.([]Order)
			clearedCh <- cleared
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