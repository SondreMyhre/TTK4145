package ordersync

import (
	"context"
	"fmt"
	"maps"
	elevio "project/elevio"
	localsingle "project/localsingleelevator"
	"time"
)

func RunWorldView(
	ctx context.Context,
	myID ElevID,

	buttonChan <-chan elevio.ButtonEvent,
	localStateChan <-chan localsingle.ElevatorState,
	clearedOrdersChan <-chan []localsingle.Order,
	netRx <-chan NetMsg,
	peerEventChan <-chan []PeerUpdate,

	netTx chan<- NetMsg,
	lightCommandChan chan<- elevio.DriverCommand,
	worldViewChan chan<- WorldviewMsg,
) error {
	state := worldviewState{
		cabRequests: make(CabCallsMap),
	}

	orderTicker := time.NewTicker(100 * time.Millisecond)

	publishWorldview := func() {
		worldview := extractWorldView(state, myID)
		worldViewChan <- worldview
	}

	for {
		var commands []command

		select {
		case <-ctx.Done():
			return nil

		case buttonEvent := <-buttonChan:
			floor := buttonEvent.Floor
			button := int(buttonEvent.Button)

			switch {
			case button == BT_CAB:
				state, commands = onCabButtonEvent(state, myID, floor)
			case button < N_HALL:
				state, commands = onHallButtonEvent(state, floor, button)
			}
			applyCommands(ctx, commands, netTx, lightCommandChan, state, myID)
			publishWorldview()

		case newLocalState := <-localStateChan:
			state.localState = newLocalState
			commands = []command{{_type: broadcastNetMessage}}
			applyCommands(ctx, commands, netTx, lightCommandChan, state, myID)
			publishWorldview()

		case cleared := <-clearedOrdersChan:
			clearedFloors, clearedButtons := convertClearedOrders(cleared)
			state, commands = onClearedOrders(state, myID, clearedFloors, clearedButtons)
			applyCommands(ctx, commands, netTx, lightCommandChan, state, myID)
			publishWorldview()

		case netMsg, ok := <-netRx:
			if !ok {
				return fmt.Errorf("Worldview: netRx closed")
			}
			state, commands = onNetMsg(state, myID, netMsg)
			applyCommands(ctx, commands, netTx, lightCommandChan, state, myID)
			publishWorldview()

		case peerEvent := <-peerEventChan:
			var newPeerList []Peer
			for _, update := range peerEvent { // state står tom
				existingState := findPeerState(state.peerList, update.ID)
				newPeerList = append(newPeerList, Peer{ID: update.ID, PeerStatus: update.PeerStatus, state: existingState})
			}
			// state, commands = onPeerEvent(state, newPeerList)
			state.peerList = newPeerList
			// applyCommands(ctx, commands, netTx, lightCommandChan, state, myID)
			publishWorldview()

		case <-orderTicker.C:
			commands = []command{{_type: broadcastNetMessage}}
			applyCommands(ctx, commands, netTx, lightCommandChan, state, myID)
			// publishWorldview()
		}
	}
}

func applyCommands(
	ctx context.Context,
	commands []command,
	netTx chan<- NetMsg,
	lightCommandChan chan<- elevio.DriverCommand,

	state worldviewState,
	myID ElevID,
) {
	for _, command := range commands {
		switch command._type {
		case broadcastNetMessage:
			cabCallsCopy := make(CabCallsMap, len(state.cabRequests))
			maps.Copy(cabCallsCopy, state.cabRequests)
			select {
			case netTx <- NetMsg{
				SenderID:        myID,
				HallOrderMatrix: state.hallOrderMatrix,
				CabCalls:        cabCallsCopy,
				SenderState:     state.localState,
			}:
			case <-ctx.Done():
				return
			}
		case setButtonLamp:
			args := command.value.(buttonLampArgs)
			select {
			case lightCommandChan <- elevio.DriverCommand{
				Type:   elevio.CommandSetButtonLamp,
				Button: elevio.ButtonType(args.Button),
				Floor:  args.Floor,
				Value:  args.Value,
			}:
			case <-ctx.Done():
				return
			}
		}

	}
}

func convertClearedOrders(orders []localsingle.Order) ([]int, []int) {
	floors := make([]int, len(orders))
	buttons := make([]int, len(orders))
	for i, order := range orders {
		floors[i] = order.Floor
		buttons[i] = int(order.Button)
	}
	return floors, buttons
}

func extractWorldView(state worldviewState, myID ElevID) WorldviewMsg {
	hallRequests := extractHallRequests(state.hallOrderMatrix)

	cabRequests := make(CabCallsMap)
	for id, calls := range state.cabRequests {
		cabRequests[id] = calls
	}

	peerStates := make(map[ElevID]localsingle.ElevatorState)
	peerStates[myID] = state.localState
	for _, peer := range state.peerList {
		peerStates[peer.ID] = peer.state

	}

	return WorldviewMsg{
		HallRequests: hallRequests,
		CabRequests:  cabRequests,
		PeerStates:   peerStates,
		Peers:        state.peerList,
	}
}

func findPeerState(peerList []Peer, id ElevID) localsingle.ElevatorState {
	for _, peer := range peerList {
		if peer.ID == id {
			return peer.state
		}
	}
	return localsingle.ElevatorState{}
}
