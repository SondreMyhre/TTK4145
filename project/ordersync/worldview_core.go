package ordersync

func onCabButtonEvent(state worldviewState, myID ElevID, floor int) (worldviewState, []command) {
	state.pendingCabCalls[floor] = true

	localCabCalls := state.cabRequests[myID]
	localCabCalls[floor] = true
	state.cabRequests[myID] = localCabCalls

	var commands []command

	if !hasAlivePeers(myID, state.peerList) {
		commands = append(commands, command{
			_type: setButtonLamp,
			value: buttonLampArgs{Floor: floor, Button: BT_CAB, Value: true},
		})
	}

	commands = append(commands, command{_type: broadcastNetMessage})
	return state, commands
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
			localCabCalls := state.cabRequests[myID]
			localCabCalls[floor] = false
			state.cabRequests[myID] = localCabCalls
			state.pendingCabCalls[floor] = false // ??
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
		state.cabRequests[senderID] = msg.CabCalls[senderID]
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
		// SPØRS OM DETTE MÅ ENDRES, BRUKE FLAGG HELLER KANSKJE??
		if remoteMyCabs, ok := msg.CabCalls[myID]; ok {
			localMyCabs := state.cabRequests[myID]
			for floor := range N_FLOORS {
				if remoteMyCabs[floor] && !localMyCabs[floor] {
					localMyCabs[floor] = true
					commands = append(commands, command{
						_type: setButtonLamp,
						value: buttonLampArgs{Floor: floor, Button: BT_CAB, Value: true},
					})
				}
			}
			state.cabRequests[myID] = localMyCabs
		}
	}

	if broadcastNeeded {
		commands = append(commands, command{_type: broadcastNetMessage})
	}

	return state, commands
}

// func onPeerEvent(state worldviewState, newPeerList []Peer) (worldviewState, []command) {
// 	for _, newPeer := range newPeerList {
// 		oldStatus := findPeerStatus(state.peerList, newPeer.ID)

// 		if oldStatus == StatusAlive && newPeer.PeerStatus == StatusDead {
// 			state.hallOrderMatrix = releaseAllConfirmed(state.hallOrderMatrix)
// 		}

// 	}
// 	return state, []command{{_type: broadcastNetMessage}}
// }

func extractHallRequests(hallOrderMatrix HallOrderMatrix) HallRequests {
	var hallRequests HallRequests
	for floor := range N_FLOORS {
		for button := range N_HALL {
			hallRequests[floor][button] = (hallOrderMatrix[floor][button].Status == Confirmed)
		}
	}
	return hallRequests
}

func hasAlivePeers(myID ElevID, peerList []Peer) bool {
	for _, peer := range peerList {
		if peer.PeerStatus == StatusAlive && peer.ID != myID {
			return true
		}
	}
	return false
}