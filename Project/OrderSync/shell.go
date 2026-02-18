package ordersync

import (
	elevio "Project/ElevIO"
	localsingle "Project/LocalSingleElevator"
	"fmt"
	"time"
)




func Run(
	buttonChan <-chan elevio.ButtonEvent,
	floorChan <-chan int,	// Trenger nok ikke
	localStateChan <-chan localsingle.LocalSingleElevator,
	clearedOrdersChan <-chan []localsingle.Order,
	rx <-chan NetMsg,
	peerEventChan <-chan []peermonitor.Peer,

	localOrderChan chan<- elevio.ButtonEvent,
	tx chan<- []NetMsg,
	lightCommandChan chan<- elevio.DriverCommand  // Muligens kun sende state og ikke hele elevator
) {
	var hallOrderMatrix OrderMatrix
	var localState LocalState
	var cabCalls [N_FLOORS]bool
 
	heartbeatTicker := time.NewTicker(100 * time.Millisecond)


	for {
		select {
		case buttonEvent := <-buttonChan:
			btn := elevioToButtonType(buttonEvent.Button)

			switch(buttonEvent) {
			case elevio.BT_Cab:
				commands = onCabButtonEvent(buttonEvent)
				
				
			case elevio.BT_HallUp || elevio.BT_HallDown:
				newOrderMatrix, commands = onHallButtonEvent(buttonEvent)
				hallOrderMatrix = newOrderMatrix
			}

		case newLocalState := <-localStateChan:
			commands = onNewLocalState(newLocalState)
			localState = newLocalState

			switch(newLocalState.behaviour) {
			case Idle:
				orderMatrixEntry = findOrder(hallOrderMatrix, peerList)
				if orderMatrixEntry != nil {
					commands = claimOrder(orderMatrixEntry)
					localOrderChan <- orderToElevioButtonEvent(orderMatrixEntry)
				}

			case DoorOpen:
				orderMatrixEntry = findOrder(hallOrderMatrix, peerList)
				if orderMatrixEntry != nil {
					commands = claimOrder(orderMatrixEntry)
					localOrderChan <- orderToElevioButtonEvent(orderMatrixEntry)
				}
			}

		case netMsg := <-rx:
			newOrderMatrix, commands = onNetMsg(hallOrderMatrix, netMsg)
			hallOrderMatrix = newOrderMatrix

		case peerEvent := <-peerEventChan:
			newOrderMatrix, newPeerList := onPeerEvent(hallOrderMatrix, peerList, peerEvent)
			hallOrderMatrix, peerList = newOrderMatrix, newPeerList

		case <-heartBeatTicker.C:
			netMsg := buildHeartbeat(hallOrderMatrix, localState) // Perhaps peerList also
			commands = sendHeartbeat(netMsg)
		}

		executeCommands(commands, localOrderChan, tx, lightCommandChan)
	}
}

func executeCommands( // Kanskje det er rotete å ha den slik når det ikke kjøres som en egen goroutine, og heller eksplisitt execute commands i hver case i Run()?
	commands []command,
	localOrderChan chan<- elevio.ButtonEvent,
	tx chan<- []NetMsg,
	lightCommandChan chan<- elevio.DriverCommand,
) {
	for _, command := range commands {
		switch command._type {
		case sendOrderToLocal:
			localOrderChan <- command.value.(elevio.ButtonEvent)
		case sendNetMsg:
			tx <- command.value.(NetMsg)
		case setButtonLamp:
			args := command.value.(buttonLampArgs)
			driverCommandChan <- elevio.DriverCommand{Type: elevio.CommandSetButtonLamp, Button: buttonTypeToElevio(args.Btn), Floor: args.Floor, Value: args.Value}
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
