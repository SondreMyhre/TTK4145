package ordersync

import (
	elevio "Project/elevio"
	localsingle "Project/localsingleelevator"
	peermonitor "Project/peermonitor"
)

func onCabButtonEvent(buttonEvent elevio.ButtonEvent) []command {
	return nil
}

func onHallButtonEvent(buttonEvent elevio.ButtonEvent) []command {
	return nil
}

func (m *HallOrderMatrix) onHallButtonEvent(buttonEvent elevio.ButtonEvent){
	floor := buttonEvent.Floor
	btn := buttonEvent.Button

	m[floor][btn].orderStatus = pending
	m[floor][btn].assignedElevator = 0
	m[floor][btn].version += 1
}

func onNewLocalState(newLocalState localsingle.LocalSingleElevator) []command {
	return nil
}

func (state *localState) updateLocalState(e localsingle.LocalSingleElevator) {

}

func findOrder(m HallOrderMatrix, pl []peer) orderMatrixEntry {
	return orderMatrixEntry{}
}

func claimOrder(o orderMatrixEntry) []command {
	return nil
}

func orderToElevioButtonEvent(o orderMatrixEntry) elevio.ButtonEvent {
	return elevio.ButtonEvent{}
}

func onNetMsg(h HallOrderMatrix, msg NetMsg) []command {
	return nil
}

func (h *HallOrderMatrix) onNetMsg(msg NetMsg) {

}

func onPeerEvent(h HallOrderMatrix, pl []peer, peerEvent []peermonitor.Peer) (ha HallOrderMatrix, pla []peer) {
	return HallOrderMatrix{}, []peer{}
}

func buildHeartbeat(h HallOrderMatrix, l localState) NetMsg {
	return NetMsg{}
}

func sendHeartbeat(msg NetMsg) []command {
	return nil
}
