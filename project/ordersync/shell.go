package ordersync

import (
	"maps"
	"time"
	elevio "project/elevio"
	localsingle "project/localsingleelevator"
)

func Run(
	myID ElevID,

	buttonChan <-chan elevio.ButtonEvent,
	localStateChan <-chan localsingle.ElevatorState,
	clearedOrdersChan <-chan []localsingle.Order,
	rx <-chan NetMsg,
	peerEventChan <-chan []Peer,

	localOrderChan chan<- elevio.ButtonEvent,
	tx chan<- NetMsg,
	lightCommandChan chan<- elevio.DriverCommand,
) {
	var hallOrderMatrix HallOrderMatrix
	var localState LocalState
	cabCalls := make(CabCallsMap)
	var pendingCabCalls [N_FLOORS]bool
	var peerList []Peer

	heartbeatTicker := time.NewTicker(100 * time.Millisecond)

	for {
		var commands []command

		select {
		case buttonEvent := <-buttonChan:
			switch buttonEvent.Button {
			case elevio.BT_Cab:
				cabCalls, pendingCabCalls, commands = onCabButtonEvent(cabCalls, pendingCabCalls, myID, buttonEvent)
				executeCommands(commands, localOrderChan, tx, lightCommandChan, hallOrderMatrix, cabCalls, myID, localState)

			case elevio.BT_HallUp, elevio.BT_HallDown:
				hallOrderMatrix, commands = onHallButtonEvent(hallOrderMatrix, buttonEvent)
				executeCommands(commands, localOrderChan, tx, lightCommandChan, hallOrderMatrix, cabCalls, myID, localState)
			}

		case newLocalState := <-localStateChan:
			hallOrderMatrix, localState, commands = onNewLocalState(hallOrderMatrix, peerList, myID, cabCalls, newLocalState)
			executeCommands(commands, localOrderChan, tx, lightCommandChan, hallOrderMatrix, cabCalls, myID, localState)

		case cleared := <-clearedOrdersChan:
			hallOrderMatrix, cabCalls, commands = onClearedOrders(hallOrderMatrix, cabCalls, myID, cleared)
			executeCommands(commands, localOrderChan, tx, lightCommandChan, hallOrderMatrix, cabCalls, myID, localState)

		case netMsg := <-rx:
			hallOrderMatrix, cabCalls, pendingCabCalls, commands = onNetMsg(hallOrderMatrix, cabCalls, myID, pendingCabCalls, peerList, netMsg)
			executeCommands(commands, localOrderChan, tx, lightCommandChan, hallOrderMatrix, cabCalls, myID, localState)

		case peerEvent := <-peerEventChan:
			hallOrderMatrix, peerList, commands = onPeerEvent(hallOrderMatrix, peerList, peerEvent)
			executeCommands(commands, localOrderChan, tx, lightCommandChan, hallOrderMatrix, cabCalls, myID, localState)

		case <-heartbeatTicker.C:
			commands = onHeartbeatTick()
			executeCommands(commands, localOrderChan, tx, lightCommandChan, hallOrderMatrix, cabCalls, myID, localState)
		}
	}
}

func executeCommands(
	commands []command,
	localOrderChan chan<- elevio.ButtonEvent,
	tx chan<- NetMsg,
	lightCommandChan chan<- elevio.DriverCommand,

	hallOrderMatrix HallOrderMatrix,
	cabCalls CabCallsMap,
	myID ElevID,
	localState LocalState,
) {
	for _, command := range commands {
		switch command._type {
		case sendOrderToLocal:
			localOrderChan <- command.value.(elevio.ButtonEvent)
		case broadcastNetMessage:
			cabCallsCopy := maps.Clone(cabCalls)
			tx <- NetMsg{
				SenderID:        myID,
				HallOrderMatrix: hallOrderMatrix,
				CabCalls:        cabCallsCopy,
				SenderState:     localState,
			}
		case setButtonLamp:
			args := command.value.(buttonLampArgs)
			lightCommandChan <- elevio.DriverCommand{
				Type:   elevio.CommandSetButtonLamp,
				Button: args.Button,
				Floor:  args.Floor,
				Value:  args.Value,
			}
		}

	}
}
