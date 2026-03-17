package ordersync

import (
	"context"
	"fmt"
	"maps"
	elevatorcontroller "project/elevatorcontroller"
	elevio "project/elevio"
	"time"
)

func RunWorldview(
	ctx context.Context,
	myID ElevID,

	buttonChan <-chan elevio.ButtonEvent,
	localStateChan <-chan elevatorcontroller.ElevatorState,
	clearedOrdersChan <-chan []elevatorcontroller.Order,
	netRx <-chan NetMsg,
	peerEventChan <-chan []PeerUpdate,

	netTx chan<- NetMsg,
	lightCommandChan chan<- elevio.DriverCommand,
	worldviewChan chan<- WorldviewMsg,
) error {
	state := worldviewState{
		cabRequests: make(CabCallsMap),
	}

	orderTicker := time.NewTicker(100 * time.Millisecond)

	for {
		var effects []effect

		select {
		case <-ctx.Done():
			return nil

		case buttonEvent := <-buttonChan:
			floor := buttonEvent.Floor
			button := int(buttonEvent.Button)

			switch {
			case button == BT_CAB:
				state, effects = onCabButtonEvent(state, myID, floor)
			case button < N_HALL:
				state, effects = onHallButtonEvent(state, floor, button)
			}
			applyEffects(ctx, effects, netTx, lightCommandChan, state, myID)
			publishWorldview(state, myID, worldviewChan)

		case newLocalState := <-localStateChan:
			state.localState = newLocalState
			effects = []effect{{kind: broadcastNetMessage}}
			applyEffects(ctx, effects, netTx, lightCommandChan, state, myID)
			publishWorldview(state, myID, worldviewChan)

		case cleared := <-clearedOrdersChan:
			clearedFloors, clearedButtons := convertClearedOrders(cleared)
			state, effects = onClearedOrders(state, myID, clearedFloors, clearedButtons)
			applyEffects(ctx, effects, netTx, lightCommandChan, state, myID)
			publishWorldview(state, myID, worldviewChan)

		case netMsg, ok := <-netRx:
			if !ok {
				return fmt.Errorf("Worldview: netRx closed")
			}
			state, effects = onNetMsg(state, myID, netMsg)
			applyEffects(ctx, effects, netTx, lightCommandChan, state, myID)
			publishWorldview(state, myID, worldviewChan)

		case peerEvent := <-peerEventChan:
			var newPeerList []Peer
			for _, update := range peerEvent {
				existingState := findPeerState(state.peerList, update.ID)
				newPeerList = append(newPeerList, Peer{ID: update.ID, PeerStatus: update.PeerStatus, state: existingState})
			}
			state.peerList = newPeerList
			publishWorldview(state, myID, worldviewChan)

		case <-orderTicker.C:
			effects = []effect{{kind: broadcastNetMessage}}
			applyEffects(ctx, effects, netTx, lightCommandChan, state, myID)
		}
	}
}

func applyEffects(
	ctx context.Context,
	effects []effect,
	netTx chan<- NetMsg,
	lightCommandChan chan<- elevio.DriverCommand,

	state worldviewState,
	myID ElevID,
) {
	for _, effect := range effects {
		switch effect.kind {
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
			args := effect.value.(buttonLampArgs)
			select {
			case lightCommandChan <- elevio.DriverCommand{
				Kind:   elevio.CommandSetButtonLamp,
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

	cabRequests := make(CabCallsMap)
	for id, calls := range state.cabRequests {
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