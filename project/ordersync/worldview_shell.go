package ordersync

import (
	"maps"
	elevatorcontroller "project/elevatorcontroller"
	elevio "project/elevio"
	"time"
)

func RunWorldview(
	myID ElevID,

	buttonChan <-chan elevio.ButtonEvent,
	localStateChan <-chan elevatorcontroller.ElevatorState,
	clearedOrdersChan <-chan []elevatorcontroller.Order,
	netRx <-chan NetMsg,
	peerEventChan <-chan []PeerUpdate,

	netTx chan<- NetMsg,
	lightCommandChan chan<- elevio.DriverCommand,
	worldviewChan chan<- WorldviewMsg,
) {
	state := worldviewState{
		cabCalls: CabCalls{Map: make(map[ElevID][N_FLOORS]bool), Version: 0},
	}

	netMsgTicker := time.NewTicker(netMsgTickInterval)

	for {
		var effects []effect

		select {
		case buttonEvent := <-buttonChan:
			floor := buttonEvent.Floor
			button := int(buttonEvent.Button)

			switch {
			case button == BT_CAB:
				state, effects = onCabButtonEvent(state, myID, floor)
			case button < N_HALL:
				state, effects = onHallButtonEvent(state, floor, button)
			}
			applyEffects(state, myID, effects, netTx, lightCommandChan)
			publishWorldview(state, myID, worldviewChan)

		case newLocalState := <-localStateChan:
			state.localState = newLocalState
			effects = []effect{{kind: broadcastNetMessage}}
			applyEffects(state, myID, effects, netTx, lightCommandChan)
			publishWorldview(state, myID, worldviewChan)

		case cleared := <-clearedOrdersChan:
			clearedFloors, clearedButtons := convertClearedOrders(cleared)
			state, effects = onClearedOrders(state, myID, clearedFloors, clearedButtons)
			applyEffects(state, myID, effects, netTx, lightCommandChan)
			publishWorldview(state, myID, worldviewChan)

		case netMsg := <-netRx:
			state, effects = onNetMsg(state, myID, netMsg)
			applyEffects(state, myID, effects, netTx, lightCommandChan)
			publishWorldview(state, myID, worldviewChan)

		case peerEvent := <-peerEventChan:
			var newPeerList []Peer
			for _, update := range peerEvent {
				existingState := findPeerState(state.peerList, update.ID)
				newPeerList = append(newPeerList, Peer{ID: update.ID, PeerStatus: update.PeerStatus, state: existingState})
			}
			state.peerList = newPeerList
			publishWorldview(state, myID, worldviewChan)

		case <-netMsgTicker.C:
			effects = []effect{{kind: broadcastNetMessage}}
			applyEffects(state, myID, effects, netTx, lightCommandChan)
		}
	}
}

func applyEffects(
	state worldviewState,
	myID ElevID,
	effects []effect,

	netTx chan<- NetMsg,
	lightCommandChan chan<- elevio.DriverCommand,

) {
	for _, effect := range effects {
		switch effect.kind {
		case broadcastNetMessage:
			cabCallsMapCopy := make(map[ElevID][N_FLOORS]bool, len(state.cabCalls.Map))
			maps.Copy(cabCallsMapCopy, state.cabCalls.Map)
			cabCallsCopy := CabCalls{Map: cabCallsMapCopy, Version: state.cabCalls.Version}
			netTx <- NetMsg{
				SenderID:        myID,
				HallOrderMatrix: state.hallOrderMatrix,
				CabCalls:        cabCallsCopy,
				SenderState:     state.localState,
			}
		case setButtonLamp:
			args := effect.value.(buttonLampArgs)
			lightCommandChan <- elevio.DriverCommand{
				Kind:   elevio.CommandSetButtonLamp,
				Button: elevio.ButtonType(args.Button),
				Floor:  args.Floor,
				Value:  args.Value,
			}
		}
	}
}

func convertClearedOrders(orders []elevatorcontroller.Order) ([]int, []int) {
	floors := make([]int, len(orders))
	buttons := make([]int, len(orders))
	for i, order := range orders {
		floors[i] = order.Floor
		buttons[i] = int(order.Button)
	}
	return floors, buttons
}

func publishWorldview(state worldviewState, myID ElevID, worldviewChan chan<- WorldviewMsg) {
	worldview := extractWorldview(state, myID)
	worldviewChan <- worldview
}

func extractWorldview(state worldviewState, myID ElevID) WorldviewMsg {
	hallRequests := extractHallRequests(state.hallOrderMatrix)

	cabRequests := make(map[ElevID][N_FLOORS]bool)
	for id, calls := range state.cabCalls.Map {
		cabRequests[id] = calls
	}

	peerStates := make(map[ElevID]elevatorcontroller.ElevatorState)
	peerStates[myID] = state.localState
	for _, peer := range state.peerList {
		if peer.ID != myID {
			peerStates[peer.ID] = peer.state
		}
	}

	return WorldviewMsg{
		HallRequests: hallRequests,
		CabRequests:  cabRequests,
		PeerStates:   peerStates,
		Peers:        state.peerList,
	}
}

func findPeerState(peerList []Peer, id ElevID) elevatorcontroller.ElevatorState {
	for _, peer := range peerList {
		if peer.ID == id {
			return peer.state
		}
	}
	return elevatorcontroller.ElevatorState{}
}
