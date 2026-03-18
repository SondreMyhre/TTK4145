package ordersync

func onCabButtonEvent(state worldviewState, myID ElevID, floor int) (worldviewState, []effect) {
	localCabCalls := state.cabCalls.Map[myID]
	localCabCalls[floor] = true
	state.cabCalls.Map[myID] = localCabCalls
	state.cabCalls.Version++

	var effects []effect

	if !hasAlivePeers(myID, state.peerList) {
		effects = append(effects, effect{
			kind:  setButtonLamp,
			value: buttonLampArgs{Floor: floor, Button: BT_CAB, Value: true},
		})
	}

	effects = append(effects, effect{kind: broadcastNetMessage})
	return state, effects
}

func onHallButtonEvent(state worldviewState, floor int, button int) (worldviewState, []effect) {
	entry := state.hallOrderMatrix[floor][button]

	if entry.Status == Inactive {
		entry.Status = Pending
		entry.Version++
		state.hallOrderMatrix[floor][button] = entry
	}

	return state, []effect{{kind: broadcastNetMessage}}
}

func onClearedOrders(state worldviewState, myID ElevID, clearedFloors []int, clearedButtons []int) (worldviewState, []effect) {
	var effects []effect

	if len(clearedFloors) == 0 {
		return state, effects
	}

	for i := range clearedFloors {
		floor := clearedFloors[i]
		button := clearedButtons[i]

		if button < N_HALL {
			entry := &state.hallOrderMatrix[floor][button]
			if entry.Status != Inactive {
				entry.Status = Inactive
				entry.Version++

				effects = append(effects, effect{
					kind: setButtonLamp,
					value: buttonLampArgs{
						Floor:  floor,
						Button: button,
						Value:  false,
					},
				})
			}
		} else {
			localCabCalls := state.cabCalls.Map[myID]
			localCabCalls[floor] = false
			state.cabCalls.Map[myID] = localCabCalls
			state.cabCalls.Version++
			effects = append(effects, effect{
				kind: setButtonLamp,
				value: buttonLampArgs{
					Floor:  floor,
					Button: button,
					Value:  false,
				},
			})
		}
	}
	effects = append(effects, effect{kind: broadcastNetMessage})

	return state, effects
}

func onNetMsg(state worldviewState, myID ElevID, msg NetMsg) (worldviewState, []effect) {
	var effects []effect

	senderID := msg.SenderID
	if senderID == myID {
		return state, effects
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
					effects = append(effects, effect{
						kind: setButtonLamp,
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
					effects = append(effects, effect{
						kind: setButtonLamp,
						value: buttonLampArgs{
							Floor:  floor,
							Button: button,
							Value:  true,
						},
					})
				case Confirmed:
					local.Status = Confirmed
					effects = append(effects, effect{
						kind: setButtonLamp,
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
					effects = append(effects, effect{
						kind: setButtonLamp,
						value: buttonLampArgs{
							Floor:  floor,
							Button: button,
							Value:  true,
						}})
				} else if remote.Status == Confirmed && local.Status != Confirmed {
					*local = remote
					effects = append(effects, effect{
						kind: setButtonLamp,
						value: buttonLampArgs{
							Floor:  floor,
							Button: button,
							Value:  true,
						}})
				} else if remote.Status == Confirmed {
					effects = append(effects, effect{
						kind: setButtonLamp,
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

	if msg.CabCalls.Version > state.cabCalls.Version {
		state.cabCalls = msg.CabCalls
	}

	localCabCalls := state.cabCalls.Map[myID]
	for floor := range N_FLOORS {
		if localCabCalls[floor] {
			effects = append(effects, effect{
				kind: setButtonLamp,
				value: buttonLampArgs{
					Floor:  floor,
					Button: BT_CAB,
					Value:  true,
				},
			})
		}

	}

	if broadcastNeeded {
		effects = append(effects, effect{kind: broadcastNetMessage})
	}

	return state, effects
}

func extractHallRequests(hallOrderMatrix HallOrderMatrix) [N_FLOORS][N_HALL]bool {
	var hallRequests [N_FLOORS][N_HALL]bool
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
