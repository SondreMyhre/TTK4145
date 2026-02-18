package ordersync

import (
	elevio "Project/ElevIO"
	localsingle "Project/LocalSingleElevator"
	peermonitor "Project/PeerMonitor"

	"golang.org/x/tools/go/analysis/passes/nilfunc"
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

func findOrder(m HallOrderMatrix, pl PeerList) orderMatrixEntry {
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

func onPeerEvent(h HallOrderMatrix, pl PeerList, peerEvent []peermonitor.Peer) (ha HallOrderMatrix, pla PeerList) {
	return HallOrderMatrix{}, PeerList{}
}

func buildHeartbeat(h HallOrderMatrix, l localState) NetMsg {
	return NetMsg{}
}

func sendHeartbeat(msg NetMsg) []command {
	return nil
}
