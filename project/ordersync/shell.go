package ordersync

import (
	"context"
	"maps"
	elevio "project/elevio"
	localsingle "project/localsingleelevator"
	"time"
	"fmt"
)

func Run(
	ctx context.Context,
	myID ElevID,

	buttonChan <-chan elevio.ButtonEvent,
	localStateChan <-chan localsingle.ElevatorState,
	clearedOrdersChan <-chan []localsingle.Order,
	rx <-chan NetMsg,
	peerEventChan <-chan []Peer,

	localOrderChan chan<- elevio.ButtonEvent,
	tx chan<- NetMsg,
	lightCommandChan chan<- elevio.DriverCommand,
) error {
	var hallOrderMatrix HallOrderMatrix
	var localState localsingle.ElevatorState
	cabCalls := make(CabCallsMap)
	var pendingCabCalls [N_FLOORS]bool
	var peerList []Peer

	heartbeatTicker := time.NewTicker(100 * time.Millisecond)
	defer heartbeatTicker.Stop()

	for {
		var commands []command

		select {
		case <-ctx.Done():
			return nil

		case buttonEvent := <-buttonChan:
			switch buttonEvent.Button {
			case elevio.BT_Cab:
				cabCalls, pendingCabCalls, commands = onCabButtonEvent(cabCalls, pendingCabCalls, myID, buttonEvent)
				executeCommands(ctx, commands, localOrderChan, tx, lightCommandChan, hallOrderMatrix, cabCalls, myID, localState)

			case elevio.BT_HallUp, elevio.BT_HallDown:
				hallOrderMatrix, commands = onHallButtonEvent(hallOrderMatrix, buttonEvent)
				executeCommands(ctx, commands, localOrderChan, tx, lightCommandChan, hallOrderMatrix, cabCalls, myID, localState)
			}

		case newLocalState := <-localStateChan:
			hallOrderMatrix, localState, commands = onNewLocalState(hallOrderMatrix, peerList, myID, cabCalls, newLocalState)
			executeCommands(ctx, commands, localOrderChan, tx, lightCommandChan, hallOrderMatrix, cabCalls, myID, localState)

		case cleared := <-clearedOrdersChan:
			hallOrderMatrix, cabCalls, commands = onClearedOrders(hallOrderMatrix, cabCalls, myID, cleared)
			executeCommands(ctx, commands, localOrderChan, tx, lightCommandChan, hallOrderMatrix, cabCalls, myID, localState)

		case netMsg, ok := <-rx:
				if !ok{
					return fmt.Errorf("ordersync: rx closed")
				}
			hallOrderMatrix, cabCalls, pendingCabCalls, commands = onNetMsg(hallOrderMatrix, cabCalls, myID, pendingCabCalls, peerList, netMsg)
			executeCommands(ctx, commands, localOrderChan, tx, lightCommandChan, hallOrderMatrix, cabCalls, myID, localState)

		case peerEvent := <-peerEventChan:
			hallOrderMatrix, peerList, commands = onPeerEvent(hallOrderMatrix, peerList, peerEvent)
			executeCommands(ctx, commands, localOrderChan, tx, lightCommandChan, hallOrderMatrix, cabCalls, myID, localState)

		case <-heartbeatTicker.C:
			commands = onHeartbeatTick()
			executeCommands(ctx, commands, localOrderChan, tx, lightCommandChan, hallOrderMatrix, cabCalls, myID, localState)
		}
	}
}

func executeCommands(
	ctx context.Context,
	commands []command,
	localOrderChan chan<- elevio.ButtonEvent,
	tx chan<- NetMsg,
	lightCommandChan chan<- elevio.DriverCommand,

	hallOrderMatrix HallOrderMatrix,
	cabCalls CabCallsMap,
	myID ElevID,
	localState localsingle.ElevatorState,
) {
	for _, command := range commands {
		switch command._type {
		case sendOrderToLocal:
			select{
			case localOrderChan <- command.value.(elevio.ButtonEvent):
			case <- ctx.Done():
				return
			}
		case broadcastNetMessage:
			cabCallsCopy := maps.Clone(cabCalls)
			select{
			case tx <- NetMsg{
				SenderID:        myID,
				HallOrderMatrix: hallOrderMatrix,
				CabCalls:        cabCallsCopy,
				SenderState:     localState,
			}:
			case <-ctx.Done():
				return 
			}
		case setButtonLamp:
			args := command.value.(buttonLampArgs)
			select{
			case lightCommandChan <- elevio.DriverCommand{
				Type:   elevio.CommandSetButtonLamp,
				Button: args.Button,
				Floor:  args.Floor,
				Value:  args.Value,
			}:
			case <- ctx.Done():
				return
			}
		}

	}
}
