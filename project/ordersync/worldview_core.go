package ordersync

import (
	localsingle "project/localsingleelevator"
)

func onCabButtonEvent(state worldviewState, myID ElevID, floor int) (worldviewState, []command) {
	state.pendingCabCalls[floor] = true

	localCabCalls := state.cabCalls[myID]
	localCabCalls[floor] = true
	state.cabCalls[myID] = localCabCalls

	return state, []command{{_type: broadcastNetMessage}}
}

func onHallButtonEvent(state worldviewState, floor int, button int) (worldviewState, []command) {
	entry := state.hallOrderMatrix[floor][button]

	if entry.Status == Inactive {
		entry.Status = Pending
		entry.Version++
		state.hallOrderMatrix[floor][button] = entry
	}

	return state, []command{{_type: broadcastNetMessage}}
}

func onNewLocalState(state worldviewState, myID ElevID, newLocalState interface{ GetObstructed() bool }) (worldviewState, []command) {

	var commands []command

	if state.localState.Obstructed {
		state.hallOrderMatrix = releaseAllConfirmed(state.hallOrderMatrix)
		commands = append(commands, command{_type: broadcastNetMessage})
	}

	return state, commands
}

func onClearedOrders(state worldviewState, myID ElevID, clearedFloors []int, clearedButtons []int) (worldviewState, []command) {
	var commands []command

	if len(clearedFloors) == 0 {
		return state, commands
	}

	for i := range clearedFloors {
		floor := clearedFloors[i]
		button := clearedButtons[i]

		if button < N_HALL {
			entry := &state.hallOrderMatrix[floor][button]
			if entry.Status != Inactive {
				entry.Status = Inactive
				entry.Version++

				commands = append(commands, command{
					_type: setButtonLamp,
					value: buttonLampArgs{
						Floor:  floor,
						Button: button,
						Value:  false,
					},
				})
			}
		} else {
			localCabCalls := state.cabCalls[myID]
			localCabCalls[floor] = false
			state.cabCalls[myID] = localCabCalls
			commands = append(commands, command{
				_type: setButtonLamp,
				value: buttonLampArgs{
					Floor:  floor,
					Button: button,
					Value:  false,
				},
			})
		}
	}
	commands = append(commands, command{_type: broadcastNetMessage})

	return state, commands
}

func onNetMsg(state worldviewState, myID ElevID, msg NetMsg) (worldviewState, []command) {
	var commands []command

	senderID := msg.SenderID
	if senderID == myID {
		return state, commands
	}

	broadcastNeeded := false

	for i := range state.peerList {
		if state.peerList[i].ID == senderID {
			state.peerList[i].state = msg.SenderState
			break
		}
	}

	if msg.SenderState.Obstructed {
		state.hallOrderMatrix = releaseAllConfirmed(state.hallOrderMatrix)
		broadcastNeeded = true
	}

	for floor := range N_FLOORS {
		for button := range N_HALL {
			remote := msg.HallOrderMatrix[floor][button]
			local := &state.hallOrderMatrix[floor][button]

			if remote.Version > local.Version {
				*local = remote

				switch local.Status {
				case Inactive:
					commands = append(commands, command{
						_type: setButtonLamp,
						value: buttonLampArgs{
							Floor:  floor,
							Button: button,
							Value:  false,
						},
					})
				case Pending:
					local.Status = Confirmed
					local.Version++
					broadcastNeeded = true
					commands = append(commands, command{
						_type: setButtonLamp,
						value: buttonLampArgs{
							Floor:  floor,
							Button: button,
							Value:  true,
						},
					})
				}

			} else if remote.Version == local.Version {
				if remote.Status == Pending && local.Status == Pending {
					local.Status = Confirmed
					local.Version++
					broadcastNeeded = true
					commands = append(commands, command{
						_type: setButtonLamp,
						value: buttonLampArgs{
							Floor:  floor,
							Button: button,
							Value:  true,
						}})
				} else if remote.Status == Confirmed && local.Status != Confirmed {
					*local = remote
					commands = append(commands, command{
						_type: setButtonLamp,
						value: buttonLampArgs{
							Floor:  floor,
							Button: button,
							Value:  true,
						}})
				} else if remote.Status == Confirmed {
					commands = append(commands, command{
						_type: setButtonLamp,
						value: buttonLampArgs{
							Floor:  floor,
							Button: button,
							Value:  true,
						},
					})
				}

			}
		}
	}

	if msg.CabCalls != nil {
		state.cabCalls[senderID] = msg.CabCalls[senderID]
		for i := range state.pendingCabCalls {
			if state.pendingCabCalls[i] && msg.CabCalls[myID][i] {
				state.pendingCabCalls[i] = false
				commands = append(commands, command{
					_type: setButtonLamp,
					value: buttonLampArgs{
						Floor:  i,
						Button: BT_CAB,
						Value:  true,
					},
				})
			}
		}
	}

	if broadcastNeeded {
		commands = append(commands, command{_type: broadcastNetMessage})
	}

	return state, commands
}

func onPeerEvent(state worldviewState, newPeerList []Peer) (worldviewState, []command) {
	var commands []command

	for _, newPeer := range newPeerList {
		oldStatus := findPeerStatus(state.peerList, newPeer.ID)

		if oldStatus == StatusAlive && newPeer.PeerStatus == StatusDead {
			state.hallOrderMatrix = releaseAllConfirmed(state.hallOrderMatrix)
		}

	}

	state.peerList = newPeerList

	return state, commands
}

func ExtractWorldView(state worldviewState, myID ElevID) Worldview {
	hallRequests := extractHallRequests(state.hallOrderMatrix)

	cabRequests := make(map[ElevID]CabRequests)
	for id, calls := range state.cabCalls {
		cabRequests[id] = calls
	}

	peerStates := make(map[ElevID]localsingle.ElevatorState)
	peerStates[myID] = state.localState
	for _, peer := range state.peerList {
		peerStates[peer.ID] = peer.state

	}

	return Worldview{
		HallRequests: hallRequests,
		CabRequests:  cabRequests,
		PeerStates:   peerStates,
		Peers:        state.peerList,
	}
}

func extractHallRequests(hallOrderMatrix HallOrderMatrix) HallRequests {
	var hallRequests HallRequests
	for floor := range N_FLOORS {
		for button := range N_BUTTONS {
			hallRequests[floor][button] = (hallOrderMatrix[floor][button].Status == Confirmed)
		}
	}
	return hallRequests
}

func findPeerStatus(peerList []Peer, id ElevID) PeerStatus {
	for _, peer := range peerList {
		if peer.ID == id {
			return peer.PeerStatus
		}
	}
	return PeerStatus(-1)
}

func releaseAllConfirmed(hallOrderMatrix HallOrderMatrix) HallOrderMatrix {
	for floor := range N_FLOORS {
		for btn := range N_HALL {
			entry := &hallOrderMatrix[floor][btn]
			if entry.Status == Confirmed {
				entry.Status = Pending
				entry.Version++
			}
		}
	}
	return hallOrderMatrix
}
